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
    title: 'Page Not Found',
    description: 'The page you are looking for might have been removed or is temporarily unavailable.',
    goBack: 'Go Back',
    home: 'Home',
  },
  nav: {
    home: 'Home',
    about: 'About Us',
    services: 'Our Advantages',
    projects: 'Case List',
    coreServices: 'Core Services',
    expertTeam: 'Expert Team',
    contact: 'Contact Us'
  },
  // DEPRECATED: Use 'home' pageKey config from usePublicContent hook
  hero: {
    title: 'Blotting Consultancy',
    subtitle: 'Enterprise Internal Regulatory Team'
  },
  // DEPRECATED: Use 'home' and 'about' pageKey configs from usePublicContent hook
  about: {
    title: 'About Us',
    description: 'Blotting Consultancy is a specialized regulatory consulting firm focused on novel foods regulated under China’s “Three New Foods” framework, genetically modified（GM）crops, and feed products，with particular expertise in the China’s regulatory system. We work closely with Chinese authorities and collaborate with trusted international partners to support regulatory submissions and approvals across China, the United States, and the European Union.',
    description2: 'We combine solid technical expertise with extensive regulatory experience to bridge the gap between product development and regulatory approval. With a deep understanding of both regulatory frameworks and underlying technologies, we help clients anticipate challenges, optimize strategies, and accelerate market entry.',
    description3: 'Our clients include multinational companies, innovative startups, and research institutions. We are committed to translating complex regulatory requirements into clear, actionable strategies that become a competitive advantage for our clients.',
    button: 'Contact Us Now'
  },
  // DEPRECATED: Use 'home' and 'advantages' pageKey configs from usePublicContent hook
  advantages: {
    title: 'Our Advantages',
    card1: {
      title: 'Science-Led Regulatory Strategy',
      titleEn: 'Science-Led Regulatory Strategy',
      description: 'We integrate scientific depth with regulatory expertise to align R&D, study design, and submission strategy from the start.',
    },
    card2: {
      title: 'Industry Insight, Aligned Communication',
      titleEn: 'Industry Insight, Aligned Communication',
      description: 'We are well-versed in both domestic and international regulatory frameworks and workflows, enabling us to collaborate smoothly and efficiently with clients’ R&D and management teams.',
    },
    card3: {
      title: 'Work Like Your In-House Reg Team',
      titleEn: 'Work Like Your In-House Reg Team',
      description: 'We act as an extension of your team, delivering practical solutions aligned with long-term goals.',
    },
    card4: {
      title: 'Transparent Process, Regulatory Engagement',
      titleEn: 'Transparent Process, Regulatory Engagement',
      description: 'We ensure transparency at every step and support constructive engagement with authorities when needed.',
    },
  },
  // DEPRECATED: Use 'home' and 'core-services' pageKey configs from usePublicContent hook
  coreServices: {
    title: 'Core Services',
    service1: {
      title: 'Novel Food Registration (NHC – China’s “Three New Foods” framework)',
      description: 'We support registrations for novel food ingredients, new food additive varieties, and new food-related product varieties, covering pathway strategy, strain construction evaluation, safety data/study planning, and expert communication to enable compliant market entry.',
      link: 'Learn More >>'
    },
    service2: {
      title: 'GM Organism Registration (Plants / Animals / Microorganisms)',
      description: 'We provide end-to-end support for MARA approvals, including Safety Certificates for Processing Use (imports) and Production & Application. Our services cover regulatory strategy development, feasibility assessments, study design, report preparation, dossier drafting, and coordination with review authorities, helping drive compliant and successful approvals.',
      link: 'Learn More >>'
    },
    service3: {
      title: 'GMM Registration for Feed Use',
      description: 'We support MARA submissions for GMM-derived feed materials, feed additives, and veterinary-active substances, including product registration strategies, evaluation of strain development, safety study planning, and support for regulatory review communications to ensure steady and effective project execution..',
    }
  },
  // DEPRECATED: Use 'core-services' pageKey config from usePublicContent hook
  coreServicesPage: {
    hero: {
      label: 'Service Areas',
      title: 'Global Regulatory Registration Services',
    },
    service1: {
      title: 'Novel Food Registration (China’s “Three New Foods” framework)',
      description: 'We support domestic and global clients in registering novel food ingredients, new varieties of food additives, and new varieties of food-related products with China’s National Health Commission (NHC). Our services cover regulatory pathway analysis, data and study planning recommendations, dossier preparation, and expert communication—helping clients reduce uncertainty and accelerate compliant market entry.',
    },
    service2: {
      title: 'Registration of Genetically Modified Organisms (Plants, Animals, and Microorganisms)',
      description: 'We support multinational and domestic companies in the regulatory registration of genetically modified (GM) organisms—plants, animals, and microorganisms—in China. Our services cover both major approval types administered by the Ministry of Agriculture and Rural Affairs (MARA): Safety Certificates for Processing Use (imports) and Safety Certificates for Production and Application (domestic use). We provide end-to-end support including pathway analysis, strategy development, safety dossier preparation, and review coordination.',
    },
    service3: {
      title: 'Regulatory Registration of Feed Materials and Feed Additives',
      description: 'We support clients with the regulatory registration of feed materials and feed additives, providing product registration strategies, evaluation of strain development, safety study planning, and review communication support to ensure steady and successful project implementation.',
    },
  },
  // DEPRECATED: Use 'advantages' pageKey config from usePublicContent hook
  advantagesPage: {
    hero: {
      label: 'Our Advantages',
      title: 'Technology-driven Regulatory Consulting',
    },
    block1: { title: 'Science-Led Regulatory Strategy', description: 'We have a deep understanding of regulatory requirements, backed by a strong foundation in biotechnology and food science. Beyond preparing submission materials, we are actively involved in product technical planning, guiding study design and data presentation to ensure seamless alignment between R&D and regulatory approval.' },
    block2: { title: 'Industry Insight, Aligned Communication', description: 'With over 20 years of experience in multinational regulatory affairs, we align smoothly with multinational clients through shared regulatory language and operating rhythms.' },
    block3: { title: 'Work Like Your In-House Reg Team', description: 'We gain an in-depth understanding of our clients’ products and strategies, acting as an embedded regulatory team to make practical and sustainable regulatory decisions together.' },
    block4: { title: 'Transparent Process, Regulatory Engagement', description: 'We ensure full transparency throughout the process and, when needed, establish effective communication with regulatory authorities.' },
    block5: { title: 'Navigate Regulatory Complexity with Confidence', description: 'By understanding the policy history, review logic, and cultural context behind China\'s rules, we help clients focus on what matters and move forward with clarity and confidence.' },
  },
  // DEPRECATED: Use 'cases' pageKey config from usePublicContent hook
  caseListPage: {
    hero: {
      label: 'Case List',
      title: 'Multiple First Cases in China',
    },
    case1: {
      title: 'GMM-Derived 2\'-Fucosyllactose (HMO)',
      items: 'Passed major review stages on the first attempt;\nenabled China\'s first domestically developed 2\'-FL approval.',
    },
    case2: {
      title: 'China\'s first approved GMM-Derived novel food ingredient (N-Acetylneuraminic Acid)',
      items: 'Successfully coordinated the efficient handoff between GMM safety assessment and novel food registration;\nbecoming the first GMM-Derived novel food ingredient approved in China.',
    },
    case3: {
      title: 'Recombinant Functional Ingredient – Market Breakthrough Strategy',
      items: 'Delivered a successful U.S. Self-GRAS;\nsupported the client in publishing papers demonstrating product safety and functionality;\nsimultaneously conducted regulatory intelligence monitoring to create opportunities for multi-market compliance planning.',
    },
    case4: {
      title: 'GM Sugarcane Safety Certificate (Processing Use Import)',
      items: 'Achieved regulatory approval rapidly despite no prior precedents.\nWhile standard safety certificates for GM crops typically take around seven years, this project secured approval in less than half that time.',
    },
  },
  // DEPRECATED: Use 'experts' pageKey config from usePublicContent hook
  expertTeamPage: {
    hero: {
      label: 'Expert Introduction',
      title: 'Rich Practical Experience in the Industry',
    },
    sectionTitle: 'Expert Introduction',
    experts: {
      xuetian: {
        name: 'Xue Tian',
        title: 'Founder',
        bio: 'Xue Tian holds a Master\'s degree in Food Science from The University of Queensland, Australia, and has more than ten years of experience in regulatory affairs within the food industry. He has previously served as a Regulatory Affairs Manager at several Fortune Global 500 companies, including Yihai Kerry Group and BASF, where he accumulated extensive hands-on industry experience.\n\nHe has been deeply involved in the development of multiple food regulations and standards, contributing professional expertise to the advancement of regulatory harmonization and industry standardization. He has also participated in numerous industry exchange and professional communication activities, sharing insights in the field of food compliance, and has developed in-depth knowledge and a strong command of both domestic and international food regulatory frameworks.\n\nCurrently, he focuses on providing end-to-end regulatory compliance consulting services for the food industry. Leading a professional team, he has supported a wide range of enterprises in navigating food regulatory requirements across Mainland China, Hong Kong, Taiwan, Australia, the European Union, and the United States. With a solid foundation in regulatory science and practice, he helps companies establish robust compliance strategies, enabling global market access and sustainable, high-quality growth.',
      },
      daiqiu: {
        name: 'Dai Qiu',
        title: 'Founder',
        bio: 'Dai Qiu has long been engaged in regulatory compliance work in the fields of biotechnology, food, and agriculture, with a particular focus on the registration and compliance of genetically modified organisms (GMOs) and novel foods ("Three New Foods") in China and international markets. He has more than ten years of continuous, hands-on project experience at the frontline of regulatory practice.\n\nSince China began gradually opening regulatory pathways for genetically modified microorganisms in 2021, he has remained deeply involved in practical submissions under the evolving regulatory framework, having participated in or led more than 30 safety assessment projects for genetically modified microbial strains.\n\nIn the novel food sector, at a time when the regulatory framework had only just been established and technical pathways and review requirements were still unclear, he was among the first to support clients in obtaining China\'s first approval for a novel food ingredient produced using genetically modified microorganisms—N-acetylneuraminic acid (sialic acid)—as well as the country\'s first approval for a human milk oligosaccharide (HMO). Projects he has led or participated in have achieved a first-round approval rate exceeding 85%, compared with an industry average of approximately 30%–40% during the same period.\n\nIn the field of genetically modified crops, three projects he participated in were among the first batch to receive China\'s Safety Certificates for the Production and Application of Genetically Modified Plants. In addition, within just three years, he supported clients in obtaining safety certificates for imported genetically modified sugarcane intended for processing use. From submission to approval, the import safety certification for genetically modified sugarcane was completed in less than three years, significantly faster than the typical six to seven years required for comparable projects.\n\nBeyond specific regulatory submissions, he has long been involved in the development of technical and regulatory systems, including experimental study design, data analysis strategy development, alignment of technical dossiers with regulatory standards, and planning of compliance pathways for global markets. From a regulatory compliance perspective, he has also authored multiple technical and regulatory articles for products such as recombinant human collagen, providing scientifically grounded support for product marketing activities.\n\nYears of frontline project experience, combined with prior roles at multinational top-tier enterprises, have given him a deep understanding of the differing perspectives of R&D, testing laboratories, and regulatory authorities. In practice, he is able to integrate and optimize across departments and industries, helping clients reduce redundant communication and misunderstandings, streamline domestic and international compliance processes, and ultimately improve regulatory approval efficiency.',
      },
    },
  },
  // DEPRECATED: Use 'about' pageKey config from usePublicContent hook
  aboutPage: {
    hero: {
      label: 'About Us',
      title: 'SPECIALIZING IN FOOD AND BIOTECHNOLOGY',
    },
    companyProfile: {
      title: 'COMPANY PROFILE',
      description: "Blotting Consultancy is a leading regulatory consulting firm specializing in novel foods regulated under China’s “Three New Foods” framework, genetically modified（GM）organisms (plants, animals, and microorganisms), feed materials, and feed additives. We have deep expertise in navigating China’s regulatory landscape and maintain close working relationships with relevant authorities. For projects involving the United States and the European Union, we collaborate with internationally recognized partners to deliver globally aligned, region-specific strategies.",
    },
    section2: {
      description: "We don’t just compile dossiers — we engineer regulatory success. Our edge lies in the rare ability to bridge science and policy: we understand not only the regulations, but also the technologies and innovations driving them. This dual insight allows us to anticipate challenges, shape smarter strategies, and accelerate approvals. Fluent in cross-cultural communication and deeply embedded in the food and biotech sectors, we help clients transform regulatory complexity into competitive advantage.",
    },
    section3: {
      description: "We have supported clients in securing approvals for multiple first-of-their-kind products and provide long-term regulatory strategy support. By consistently delivering at critical milestones, we build lasting trust and credibility with our clients.",
    },
  },
  // DEPRECATED: Use 'contact' pageKey config from usePublicContent hook
  contactPage: {
    hero: {
      title: 'Contact Us',
      subtitle: 'Please follow us',
    },
    form: {
      title: 'Contact our experts',
      subtitle: 'We will all support you.',
      nameLabel: 'Name',
      emailLabel: 'Email',
      messageLabel: 'Your message',
      namePlaceholder: 'Your name',
      emailPlaceholder: 'Your email address',
      messagePlaceholder: 'Your message',
      submit: 'Submit',
    },
    contact: {
      phone: 'Mr. Xue 15910769614',
      address: 'Address: No. 115, 1/F, Building 3, Courtyard 9, West Huilongguan Street, Changping District, Beijing',
    },
  },
  // DEPRECATED: Use 'global' pageKey config from usePublicContent hook (partially)
  // Note: Some footer content may still be in transition to config-driven rendering
  footer: {
    address: 'Address: Building 3, Floor 1, Unit 115, No. 9 West Huilongguan Street, Changping District, Beijing',
    phone: 'Phone: +86 159 1076 9614',
    companySection: 'Company',
    aboutUs: 'About Us',
    ourTeam: 'Professional Team',
    caseStudies: 'Case Studies',
    contactUs: 'Contact Us',
    servicesSection: 'Services',
    regulatoryConsulting: 'Regulatory Consulting "Three New Foods" Registration',
    gmoRegistration: 'GMO Registration (Plant/Animal/Microorganism)',
    complianceTraining: 'Compliance Training and Continuous Regulatory Support',
    copyright: 'Copyright © 2018-2025  All Rights Reserved  Blotting Consultancy (Beijing) Co., Ltd.  ICP'
  }
};
