/**
 * DEPRECATION NOTICE (as of FE-108):
 *
 * Most keys in this file are OBSOLETE and retained only for backward compatibility.
 * All public page content is now managed via backend CMS and fetched using:
 * - usePublicContent hook from @/hooks/usePublicContent
 * - Page configs from backend /public/content/{pageKey} API
 *
 * Deprecated key groups (DO NOT USE in new code):
 * - hero.* → Replaced by 'home' pageKey config
 * - about.* → Replaced by 'home' and 'about' pageKey configs
 * - advantages.* → Replaced by 'home' and 'advantages' pageKey configs
 * - coreServices.* → Replaced by 'home' and 'core-services' pageKey configs
 * - coreServicesPage.* → Replaced by 'core-services' pageKey config
 * - advantagesPage.* → Replaced by 'advantages' pageKey config
 * - caseListPage.* → Replaced by 'cases' pageKey config
 * - expertTeamPage.* → Replaced by 'experts' pageKey config
 * - aboutPage.* → Replaced by 'about' pageKey config
 * - contactPage.* → Replaced by 'contact' pageKey config
 * - footer.* → Replaced by 'global' pageKey config (partially)
 *
 * Active keys (still in use):
 * - nav.* → May be used by legacy components (migration to 'global' config in progress)
 *
 * See src/FRONTEND_RENDERING.md for migration guide and usage patterns.
 */

export const common = {
  notFound: {
    title: '页面未找到',
    description: '您访问的页面可能已被移除或暂时不可用。',
    goBack: '返回上页',
    home: '返回首页',
  },
  nav: {
    home: '首页',
    about: '关于我们',
    services: '我们的优势',
    projects: '案例清单',
    coreServices: '核心服务',
    expertTeam: '专家团队',
    contact: '联系我们'
  },
  // DEPRECATED: Use 'home' pageKey config from usePublicContent hook
  hero: {
    title: '印迹法规咨询',
    subtitle: '企业内设型注册法规团队'
  },
  // DEPRECATED: Use 'home' and 'about' pageKey configs from usePublicContent hook
  about: {
    title: '关于我们',
    description: '印迹法规咨询公司是一家专注于食品与生物技术领域的专业法规注册咨询公司，核心业务涵盖“三新食品”、转基因作物和饲料注册。我们与中国监管机构保持密切沟通，并携手国际合作伙伴，开展中国、美国和欧盟的法规注册项目。',
    description2: '我们将扎实的技术背景和资深的法规经验充分融合，打通产品上市与法规批准之间的壁垒。凭借对法规体系与产品技术的深刻理解，公司团队帮助客户预判风险、优化策略、加速获批。',
    description3: '我们的客户涵盖跨国企业、初创公司和科研机构。我们致力于将复杂的监管要求精准地转化为客户的竞争优势。',
    button: '立即联系我们'
  },
  // DEPRECATED: Use 'home' and 'advantages' pageKey configs from usePublicContent hook
  advantages: {
    title: '我们的优势',
    card1: {
      title: '技术驱动的法规咨询',
      titleEn: 'Science-Led Regulatory Strategy',
      description: '依托坚实的食品与生物技术基础，提供从技术方案、试验设计到材料撰写的全流程服务，高效贯通企业研发与法规注册环节。',
    },
    card2: {
      title: '行业视角，高效协作',
      titleEn: 'Industry Insight, Aligned Communication',
      description: '熟悉国内外法规监管语言与工作逻辑，能够与客户研发管理团队建立顺畅、高效的专业协作关系。',
    },
    card3: {
      title: '像企业内设型团队一样工作',
      titleEn: 'Work Like Your In-House Reg Team',
      description: '深入理解客户的产品与战略，作为客户的内设型法规团队，共同做出可落地、可持续的法规决策。',
    },
    card4: {
      title: '透明的工作流程，主动对接监管',
      titleEn: 'Transparent Process, Regulatory Engagement',
      description: '全过程透明推进，必要时与监管机构建立有效沟通。',
    },
  },
  // DEPRECATED: Use 'home' and 'core-services' pageKey configs from usePublicContent hook
  coreServices: {
    title: '核心服务',
    service1: {
      title: '卫健委"三新食品"注册',
      description: '覆盖新食品原料、食品添加剂新品种及食品相关产品新品种，提供法规战略研判、协助试验设计、撰写申报材料并与专家沟通，助力产品快速合规上市。',
      link: '了解更多 >>'
    },
    service2: {
      title: '转基因生物注册（植物/动物/微生物）',
      description: '开展转基因安全证书申报，包括进口加工原料安全证书与生产应用安全证书。提供申报策略、可行性分析、试验设计、报告撰写、申报书撰写和审评沟通协调服务，推动产品合规批准.',
      link: '了解更多 >>'
    },
    service3: {
      title: '饲料领域注册',
      description: '开展饲料原料和饲料添加剂法规申报，提供产品注册申报策略、菌株构建评估、安全性试验规划与审评沟通支持，稳步推进项目落地。',
      link: '了解更多 >>'
    }
  },
  // DEPRECATED: Use 'core-services' pageKey config from usePublicContent hook
  coreServicesPage: {
    hero: {
      label: '服务领域',
      title: '全球法规注册服务',
    },
    service1: {
      title: '卫健委"三新食品"注册',
      description: '我们为国内外客户提供新食品原料、食品添加剂新品种、食品相关产品新品种的注册支持，覆盖国家卫生健康委的全流程审批要求。服务内容包括法规路径判断、资料准备建议、试验策略咨询、申报材料撰写及专家评审沟通，帮助客户识别关键风险点、优化技术方案，并推动产品合规上市。',
    },
    service2: {
      title: '转基因生物注册（植物/动物/微生物）',
      description: '我们为跨国公司和中国企业提供转基因生物（植物、动物和微生物）的法规注册服务，包含“进口用作加工原料的安全证书”与“生产应用安全证书”两大类。服务包括法规路径判断、申报策略制定、安全性评价资料准备、审评沟通协调等，支持产品合规进入中国市场。',
    },
    service3: {
      title: '饲料原料和饲料添加剂注册',
      description: '我们为客户提供饲料原料和饲料添加剂的法规注册申报服务，服务内容包括产品注册申报策略制定、菌株构建评估、安全性试验规划与审评沟通支持，以稳步推进项目落地。',
    },
  },
  // DEPRECATED: Use 'advantages' pageKey config from usePublicContent hook
  advantagesPage: {
    hero: {
      label: '我们的优势',
      title: '技术驱动的法规咨询',
    },
    block1: { title: '技术驱动的法规咨询', description: '我们深度理解监管政策要求，同时具备坚实的生物技术和食品科学行业背景。除材料撰写外，我们深度参与产品技术方案的设计，指导试验设计与数据呈现，确保研发与注册的逻辑紧密衔接。' },
    block2: { title: '行业视角驱动的高效协作', description: '凭借二十余年的跨国企业法规注册经验，我们更熟悉国际监管语言与专业思维方式，能与客户建立高效、同频的专业协作。' },
    block3: { title: '像企业内设型团队一样工作', description: '深入理解客户的产品与战略，作为客户的内设型法规团队，共同做出可落地、可持续的法规决策。' },
    block4: { title: '过程透明，主动对接监管', description: '全过程透明推进，必要时与监管机构建立有效沟通。' },
    block5: { title: '化繁为简，掌控复杂合规', description: '基于对政策背景、审评逻辑及中国特有监管语境的把握，我们能帮助客户精准定位关键决策点，在复杂环境中做出清晰判断，并稳健推动项目进程。' },
  },
  // DEPRECATED: Use 'cases' pageKey config from usePublicContent hook
  caseListPage: {
    hero: {
      label: '案例清单',
      title: '中国多个首例',
    },
    case1: {
      title: 'GMM 生产的 2\'-岩藻糖基乳糖（HMO）项目',
      items: '关键评审环节一次性通过；实现中国首个国产HMO（2\'-FL）获批。',
    },
    case2: {
      title: '中国首例获批的遗传修饰微生物生产的新食品原料项目（N-乙酰神经氨酸）',
      items: '完成 GMM 安评与新食品原料申报的高效衔接推进；\n成为中国首个获批的 GMM 生产新食品原料。',
    },
    case3: {
      title: '重组功能蛋白原料市场破局项目',
      items: '完成美国GRAS；\n为客户发表论文，阐述产品安全性与功效；\n同步开展政策情报监测，为后续多市场合规布局预留窗口。（给出引用论文）',
    },
    case4: {
      title: '转基因甘蔗进口用作加工原料安全证书',
      items: '在没有批准先例的情况下，快速完成法规批准。\n常规转基因植物安全证书需7年左右，该项目仅用一半不到的时间即获得批准。',
    },
  },
  // DEPRECATED: Use 'experts' pageKey config from usePublicContent hook
  expertTeamPage: {
    hero: {
      label: '专家介绍',
      title: '丰富的行业实战经验',
    },
    sectionTitle: '专家介绍',
    experts: {
      xuetian: {
        name: '薛天',
        title: '创始人',
        bio: '毕业于澳大利亚昆士兰大学，食品科学硕士研究生学历，拥有超过十年的食品行业法规事务经验，曾任职于益海嘉里集团、巴斯夫等世界500强企业，担任法规事务经理，积累了丰富的行业实战经验。\n\n深度参与多项食品法规与标准的制定工作，为推动行业规范化发展贡献专业力量。曾参与多地行业交流活动，分享食品合规领域专业见解，对国内外食品法规体系有着深入研究与精准把握。\n\n现专注于为企业提供覆盖全链条的食品行业合规咨询服务，带领团队为众多企业提供涵盖中国大陆、香港、台湾、澳大利亚、欧盟及美国等国家或地区的食品合规解决方案，以深厚的法规积淀为企业业务拓展筑牢合规根基，助力企业实现全球化合规布局与高质量发展。',
      },
      daiqiu: {
        name: '戴秋',
        title: '创始人',
        bio: '戴秋，长期从事生物技术与食品、农业法规合规工作，专注于转基因生物、三新食品在中国及国际市场的注册与合规事务，拥有十余年持续的一线项目经验。\n\n自 2021 年我国逐步放开转基因微生物相关申报路径以来，他持续参与相关制度下的实际申报工作，累计参与和主导 30 余个转基因微生物菌株安全性评价项目。\n\n在"三新食品"领域，监管制度刚刚放开，技术路径和审评要求不清晰的背景下，他率先支持客户获得全国首个转基因微生物生产的新食品原料唾液酸的批准，和全国首个母乳低聚糖（HMO）的批准。他主导、参与的项目一次性通过率超过 85%，而同期业内通常仅为 30%–40%。\n\n在转基因农作物领域，第一批获得"中国转基因植物生产应用安全证书"的项目中有3个来自他参与的项目。此外，他仅用3年便助力客户获得了进口转基因甘蔗加工原料的安全证书。转基因甘蔗的进口用作加工原料的安全证书在不足三年的时间内完成从递交到批准，明显快于同类项目常规所需的6至7年。\n\n除具体申报工作外，他还长期参与技术体系建设，涉及实验方案设计、数据分析方案设计，以及技术材料与法规标准之间的路径规划、全球市场的合规路径布局。同时，他从合规角度出发，为客户的重组人源胶原蛋白等产品撰写了多篇技术与法规相关文章，为产品的市场营销提供了坚实的科学支撑。\n\n凭借多年一线经验及在跨国百强企业的工作履历，他深谙研发、检测、监管等各方的不同逻辑与诉求。在实际工作中，他能够实现跨部门、跨行业的高效整合优化，帮助客户减少重复沟通和理解偏差，打通海内外合规路径，从而提升整体获批效率。',
      },
    },
  },
  // DEPRECATED: Use 'about' pageKey config from usePublicContent hook
  aboutPage: {
    hero: {
      label: '关于我们',
      title: '专注食品与生物技术法规',
    },
    companyProfile: {
      title: '公司简介',
      description: '印迹安合是一家专注于食品与生物技术领域的专业法规咨询公司，擅长“三新食品”、转基因生物（植物、动物和微生物）、饲料原料和饲料添加剂的注册申报。我们深耕中国的监管法规体系，与相关主管部门保持紧密沟通与合作。在美国和欧盟的申报项目中，我们与国际知名专家协同合作，确保为客户提供全球一致、区域精准的注册策略。',
    },
    section2: {
      description: '我们不仅限于撰写申报材料，更致力于实现法规突破。我们的优势在于科学与政策的深度融合——不仅了解监管规则，更理解产品背后的技术原理与产业趋势。正因如此，我们能够预判挑战、制定更具前瞻性的策略，并加速获批进程。凭借跨文化沟通能力以及对食品与生物技术行业的深度理解，我们帮助客户将复杂的监管体系转化为切实的竞争优势。',
    },
    section3: {
      description: '我们帮助客户获得了多项首例产品的批准，并为客户提供长期的法规申报战略支持。我们始终在关键节点交付成果，赢得信赖。',
    },
  },
  // DEPRECATED: Use 'contact' pageKey config from usePublicContent hook
  contactPage: {
    hero: {
      title: '联系我们',
      subtitle: '请关注我们',
    },
    form: {
      title: '联络我们的专家',
      subtitle: '我们都将为你提供支持。',
      nameLabel: '姓名',
      emailLabel: '邮箱',
      messageLabel: '留言',
      namePlaceholder: '您的姓名',
      emailPlaceholder: '您的邮箱地址',
      messagePlaceholder: '您的留言',
      submit: '提交',
    },
    contact: {
      phone: '薛先生 15910769614',
      address: '地址：北京市昌平区回龙观西大街9号院3号楼1层115',
    },
  },
  // DEPRECATED: Use 'global' pageKey config from usePublicContent hook (partially)
  // Note: Some footer content may still be in transition to config-driven rendering
  footer: {
    address: '地址：北京市昌平区回龙观西大街9号院3号楼1层115',
    phone: '电话：+86 159 1076 9614',
    companySection: '公司',
    aboutUs: '关于我们',
    ourTeam: '专业团队',
    caseStudies: '成功案例',
    contactUs: '联系我们',
    servicesSection: '服务',
    regulatoryConsulting: '法规咨询"三新食品"注册',
    gmoRegistration: '转基因生物注册（植物/动物/微生物）',
    complianceTraining: '合规培训和持续法规支持',
    copyright: '版权所有 © 2018-2025  保留所有权利  印迹法规（北京）咨询有限公司  京ICP备'
  }
};
