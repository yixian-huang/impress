package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"blotting-consultancy/internal/middleware"
	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
	"blotting-consultancy/internal/service"
)

type handlerHarness struct {
	handler   *Handler
	scheduler *service.SchedulerService
	articles  repository.ArticleRepository
	pages     repository.UnifiedPageRepository
}

func newHandlerHarness(t *testing.T) *handlerHarness {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.Article{},
		&model.Category{},
		&model.Tag{},
		&model.UnifiedPage{},
		&model.PageVersion{},
		&model.ScheduledPublishJob{},
	))
	articles := repository.NewGormArticleRepository(db)
	pages := repository.NewGormUnifiedPageRepository(db)
	versions := repository.NewGormPageVersionRepository(db)
	articleSvc := service.NewArticlePublicationService(articles, nil, nil).
		WithTaxonomyRepositories(
			repository.NewGormCategoryRepository(db),
			repository.NewGormTagRepository(db),
		)
	pageSvc := service.NewUnifiedPageService(pages, versions)
	schedulerSvc := service.NewSchedulerService(
		repository.NewGormScheduledPublishJobRepository(db),
		articleSvc,
		pageSvc,
	)
	return &handlerHarness{
		handler:   NewHandler(schedulerSvc),
		scheduler: schedulerSvc,
		articles:  articles,
		pages:     pages,
	}
}

func routerForUser(handler *Handler, user *model.User) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("rbac_user", user)
		c.Set(string(middleware.UserContextKey), &middleware.UserContext{
			UserID:   user.ID,
			Username: user.Username,
			Role:     model.RoleAdmin,
		})
		c.Next()
	})
	router.GET("/admin/scheduled-publications", handler.List)
	router.POST("/admin/scheduled-publications", handler.Schedule)
	router.PUT("/admin/scheduled-publications/:id", handler.Reschedule)
	router.DELETE("/admin/scheduled-publications/:id", handler.Cancel)
	router.POST("/admin/scheduled-publications/:id/retry", handler.Retry)
	return router
}

func userWithPermissions(id uint, permissions ...model.Permission) *model.User {
	return &model.User{
		ID:       id,
		Username: fmt.Sprintf("user-%d", id),
		Role:     model.RoleAdmin,
		UserRoles: []model.UserRole{{
			Role: model.RBACRole{Permissions: permissions},
		}},
	}
}

func performJSON(
	t *testing.T,
	router *gin.Engine,
	method string,
	path string,
	payload interface{},
) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	if payload != nil {
		require.NoError(t, json.NewEncoder(&body).Encode(payload))
	}
	request := httptest.NewRequest(method, path, &body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestScheduledPublicationHandlerUsesNeutralContractAndDeleteCancellation(t *testing.T) {
	h := newHandlerHarness(t)
	ctx := context.Background()
	article := &model.Article{
		Slug:          "contract-article",
		ZhTitle:       "Contract Article",
		Status:        model.ArticleStatusDraft,
		AllowComments: true,
	}
	require.NoError(t, h.articles.Create(ctx, article))

	router := routerForUser(h.handler, userWithPermissions(
		1,
		model.Permission{Resource: "articles", Action: "publish"},
	))
	scheduledAt := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	response := performJSON(t, router, http.MethodPost, "/admin/scheduled-publications", map[string]interface{}{
		"resourceType": "article",
		"resourceId":   article.ID,
		"scheduledAt":  scheduledAt,
	})
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())

	var created scheduledPublicationResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &created))
	require.Equal(t, model.ScheduledContentArticle, created.ResourceType)
	require.Equal(t, article.ID, created.ResourceID)
	require.Equal(t, "Contract Article", created.Title)
	require.Equal(t, "contract-article", created.Slug)
	require.Equal(t, model.ScheduledJobPending, created.Status)

	cancelResponse := performJSON(
		t,
		router,
		http.MethodDelete,
		fmt.Sprintf("/admin/scheduled-publications/%d", created.ID),
		nil,
	)
	require.Equal(t, http.StatusOK, cancelResponse.Code, cancelResponse.Body.String())
	var cancelled scheduledPublicationResponse
	require.NoError(t, json.Unmarshal(cancelResponse.Body.Bytes(), &cancelled))
	require.Equal(t, model.ScheduledJobCancelled, cancelled.Status)
	require.NotNil(t, cancelled.CompletedAt)
}

func TestScheduledPublicationHandlerFiltersQueueByResourcePermission(t *testing.T) {
	h := newHandlerHarness(t)
	ctx := context.Background()
	scheduledAt := time.Now().Add(time.Hour)

	article := &model.Article{
		Slug:          "private-queue-article",
		ZhTitle:       "Private Queue Article",
		Status:        model.ArticleStatusDraft,
		AllowComments: true,
	}
	require.NoError(t, h.articles.Create(ctx, article))
	page := &model.UnifiedPage{
		Slug:         "visible-queue-page",
		ZhTitle:      "Visible Queue Page",
		Mode:         model.PageModeComposable,
		DraftConfig:  model.JSONMap{"sections": []interface{}{}},
		DraftVersion: 1,
		Status:       "draft",
	}
	require.NoError(t, h.pages.Create(ctx, page))
	_, err := h.scheduler.Schedule(
		ctx,
		model.ScheduledContentArticle,
		article.ID,
		scheduledAt,
		nil,
		nil,
		1,
	)
	require.NoError(t, err)
	expectedVersion := page.DraftVersion
	_, err = h.scheduler.Schedule(
		ctx,
		model.ScheduledContentPage,
		page.ID,
		scheduledAt,
		&expectedVersion,
		nil,
		1,
	)
	require.NoError(t, err)

	router := routerForUser(h.handler, userWithPermissions(
		2,
		model.Permission{Resource: "pages", Action: "publish"},
	))
	listResponse := performJSON(t, router, http.MethodGet, "/admin/scheduled-publications", nil)
	require.Equal(t, http.StatusOK, listResponse.Code, listResponse.Body.String())
	var list struct {
		Items []scheduledPublicationResponse `json:"items"`
		Total int                            `json:"total"`
	}
	require.NoError(t, json.Unmarshal(listResponse.Body.Bytes(), &list))
	require.Equal(t, 1, list.Total)
	require.Len(t, list.Items, 1)
	require.Equal(t, model.ScheduledContentPage, list.Items[0].ResourceType)

	forbidden := performJSON(t, router, http.MethodPost, "/admin/scheduled-publications", map[string]interface{}{
		"resourceType": "article",
		"resourceId":   article.ID,
		"scheduledAt":  time.Now().Add(2 * time.Hour),
	})
	require.Equal(t, http.StatusForbidden, forbidden.Code, forbidden.Body.String())
}
