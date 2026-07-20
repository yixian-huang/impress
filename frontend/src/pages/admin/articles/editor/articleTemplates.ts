/**
 * Built-in article structure templates for the admin editor.
 * Bodies are HTML (TipTap-friendly). Markdown conversion uses htmlToMarkdown when needed.
 */

export type ArticleTemplateId = "blank" | "tutorial" | "news" | "case-study";

export interface ArticleTemplate {
  id: ArticleTemplateId;
  name: string;
  description: string;
  zhTitle: string;
  enTitle: string;
  zhBody: string;
  enBody: string;
}

export const ARTICLE_TEMPLATES: ArticleTemplate[] = [
  {
    id: "blank",
    name: "空白",
    description: "从空文档开始",
    zhTitle: "",
    enTitle: "",
    zhBody: "<p></p>",
    enBody: "<p></p>",
  },
  {
    id: "tutorial",
    name: "教程 / How-to",
    description: "目标 → 步骤 → 小结，适合操作指南",
    zhTitle: "如何…",
    enTitle: "How to…",
    zhBody: [
      "<h2>你将学到什么</h2>",
      "<p>用一两句话说明读者完成这篇教程后能做到什么。</p>",
      "<h2>前提条件</h2>",
      "<ul><li>条件一</li><li>条件二</li></ul>",
      "<h2>步骤</h2>",
      "<h3>步骤 1</h3><p>描述操作，并附上截图或代码。</p>",
      "<h3>步骤 2</h3><p>继续下一步。</p>",
      "<h3>步骤 3</h3><p>完成关键动作。</p>",
      "<h2>常见问题</h2>",
      "<p><strong>Q：…？</strong><br>A：…</p>",
      "<h2>小结</h2>",
      "<p>回顾要点，并给出下一步阅读建议。</p>",
    ].join(""),
    enBody: [
      "<h2>What you'll learn</h2>",
      "<p>In one or two sentences, describe the outcome.</p>",
      "<h2>Prerequisites</h2>",
      "<ul><li>Requirement one</li><li>Requirement two</li></ul>",
      "<h2>Steps</h2>",
      "<h3>Step 1</h3><p>Describe the action.</p>",
      "<h3>Step 2</h3><p>Continue.</p>",
      "<h3>Step 3</h3><p>Finish the flow.</p>",
      "<h2>FAQ</h2>",
      "<p><strong>Q: …?</strong><br>A: …</p>",
      "<h2>Summary</h2>",
      "<p>Recap key points and suggest next reads.</p>",
    ].join(""),
  },
  {
    id: "news",
    name: "资讯 / News",
    description: "导语 + 要点 + 详情，适合发布公告",
    zhTitle: "【资讯】",
    enTitle: "[News]",
    zhBody: [
      "<p><strong>导语：</strong>用一句话概括最重要的信息。</p>",
      "<h2>要点</h2>",
      "<ul><li>要点一</li><li>要点二</li><li>要点三</li></ul>",
      "<h2>详情</h2>",
      "<p>展开背景、时间线与影响。</p>",
      "<h2>相关链接</h2>",
      "<ul><li><a href=\"#\">相关文档</a></li></ul>",
    ].join(""),
    enBody: [
      "<p><strong>Lead:</strong> One sentence with the key news.</p>",
      "<h2>Highlights</h2>",
      "<ul><li>Point one</li><li>Point two</li><li>Point three</li></ul>",
      "<h2>Details</h2>",
      "<p>Background, timeline, and impact.</p>",
      "<h2>Links</h2>",
      "<ul><li><a href=\"#\">Related docs</a></li></ul>",
    ].join(""),
  },
  {
    id: "case-study",
    name: "案例 / Case study",
    description: "背景 → 方案 → 结果，适合客户故事",
    zhTitle: "案例：",
    enTitle: "Case study:",
    zhBody: [
      "<h2>客户背景</h2>",
      "<p>行业、规模与业务场景。</p>",
      "<h2>挑战</h2>",
      "<p>对方遇到的核心问题。</p>",
      "<h2>解决方案</h2>",
      "<ol><li>方案要点一</li><li>方案要点二</li></ol>",
      "<h2>结果与数据</h2>",
      "<ul><li>指标 A：</li><li>指标 B：</li></ul>",
      "<h2>客户评价</h2>",
      "<blockquote><p>引用一句客户反馈。</p></blockquote>",
    ].join(""),
    enBody: [
      "<h2>Background</h2>",
      "<p>Industry, scale, and context.</p>",
      "<h2>Challenge</h2>",
      "<p>The core problem.</p>",
      "<h2>Solution</h2>",
      "<ol><li>Approach one</li><li>Approach two</li></ol>",
      "<h2>Results</h2>",
      "<ul><li>Metric A:</li><li>Metric B:</li></ul>",
      "<h2>Quote</h2>",
      "<blockquote><p>A short customer quote.</p></blockquote>",
    ].join(""),
  },
];

export function getArticleTemplate(id: string): ArticleTemplate | undefined {
  return ARTICLE_TEMPLATES.find((t) => t.id === id);
}
