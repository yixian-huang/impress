import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import PageHero from '../../components/feature/PageHero';
import { usePublicContent } from '@/hooks/usePublicContent';
import type { Locale } from '@/api/publicContent';
import { PublicLayout } from '@/theme/layouts';

interface MediaRef {
  url?: string;
  alt?: string;
}

interface HeroConfig {
  label?: string;
  title?: string;
  image?: MediaRef;
}

interface Expert {
  id: string;
  name?: string;
  title?: string;
  avatar?: MediaRef;
  bioParagraphs?: string[];
}

interface ExpertsPageConfig {
  hero?: HeroConfig;
  sectionTitle?: string;
  experts?: Expert[];
}

export default function ExpertsPage() {
  const { i18n } = useTranslation('common');
  const locale = (i18n.language === 'zh' || i18n.language.startsWith('zh') ? 'zh' : 'en') as Locale;

  const { loading, error, config } = usePublicContent('experts', {
    locale,
    autoNormalize: true,
  });

  const pageConfig = (config as ExpertsPageConfig) || {};
  const hero = pageConfig.hero || {};
  const experts = Array.isArray(pageConfig.experts) ? pageConfig.experts : [];

  const [activeId, setActiveId] = useState<string>(experts[0]?.id || '');

  if (loading) {
    return (
      <PublicLayout>
        <div className="min-h-screen bg-white flex items-center justify-center">
          <div className="text-gray-600">Loading...</div>
        </div>
      </PublicLayout>
    );
  }

  if (error) {
    return (
      <PublicLayout>
        <div className="min-h-screen bg-white flex items-center justify-center">
          <div className="text-red-600">Failed to load page content</div>
        </div>
      </PublicLayout>
    );
  }

  const activeExpert = experts.find((e) => e.id === activeId) || experts[0];
  const bioParagraphs = activeExpert?.bioParagraphs || [];

  return (
    <PublicLayout>
      <PageHero
        label={hero.label}
        title={hero.title}
        alt="Expert Team Hero"
        imageSrc={hero.image?.url}
      />

      {/* 专家介绍 */}
      <section className="py-12 md:py-16 lg:py-24 bg-white">
        <div className="max-w-layout mx-auto px-4 md:px-6">
          {/* 区块标题 */}
          {pageConfig.sectionTitle && (
            <div className="flex items-center mb-10 md:mb-12">
              <div className="w-[26px] h-[26px] bg-accent mr-3 flex-shrink-0 rounded-full" />
              <h2 className="text-2xl md:text-3xl font-bold text-gray-900">
                {pageConfig.sectionTitle}
              </h2>
            </div>
          )}

          {/* 专家头像 + 姓名职位 */}
          {experts.length > 0 && (
            <div className="grid grid-cols-2 gap-8 md:gap-12 max-w-2xl mx-auto mb-12 md:mb-16">
              {experts.map((expert) => (
                <div key={expert.id} className="flex flex-col items-center text-center">
                  <div className="w-32 h-32 md:w-40 md:h-40 rounded-full overflow-hidden border-2 border-gray-200 flex-shrink-0 mb-3">
                    <img
                      src={expert.avatar?.url || `/images/expert/${expert.id}.png`}
                      alt={expert.name || expert.id}
                      className="w-full h-full object-cover object-top"
                    />
                  </div>
                  {expert.name && (
                    <h3 className="text-lg md:text-xl font-bold text-primary">
                      {expert.name}
                    </h3>
                  )}
                  {expert.title && (
                    <p className="text-sm text-gray-500">
                      {expert.title}
                    </p>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* 左侧 Tab + 右侧简介正文 */}
          {experts.length > 0 && activeExpert && (
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 lg:gap-8 items-start">
              <div className="lg:col-span-3 flex lg:flex-col gap-2">
                {experts.map((expert) => (
                  <button
                    key={expert.id}
                    type="button"
                    onClick={() => setActiveId(expert.id)}
                    className={`px-4 py-3 text-left rounded-md transition-colors cursor-pointer whitespace-nowrap ${
                      activeId === expert.id
                        ? 'bg-primary text-white'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                    }`}
                  >
                    {expert.name || expert.id}
                  </button>
                ))}
              </div>
              <div className="lg:col-span-9 bg-white rounded-lg border border-gray-100 p-6 md:p-8">
                {bioParagraphs.map((para, i) => (
                  <p key={i} className="text-base text-gray-700 leading-relaxed mb-4 last:mb-0">
                    {para}
                  </p>
                ))}
              </div>
            </div>
          )}
        </div>
      </section>
    </PublicLayout>
  );
}
