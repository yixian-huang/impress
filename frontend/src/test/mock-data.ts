/**
 * Mock configurations for testing public content pages
 * These configs match the structure of backend published configs
 */

export const mockHomeConfig = {
  hero: {
    title: "Test Hero Title",
    subtitle: "Test Hero Subtitle",
    backgroundImage: "/images/hero-bg.png",
  },
  about: {
    title: "Test About Title",
    descriptions: [
      "Test About Description",
      "Test About Description 2",
      "Test About Description 3",
    ],
    button: "Learn More",
    image: "/images/about-img.png",
  },
  advantages: {
    title: "Test Advantages Title",
    subtitle: "Test Advantages Subtitle",
    cards: [
      {
        title: "Advantage 1",
        description: "Advantage 1 Description",
        image: "/images/advantage-1.png",
      },
      {
        title: "Advantage 2",
        description: "Advantage 2 Description",
        image: "/images/advantage-2.png",
      },
    ],
  },
  coreServices: {
    title: "Test Services Title",
    subtitle: "Test Services Subtitle",
    items: [
      {
        title: "Service 1",
        description: "Service 1 Description",
        image: "/images/service-1.png",
      },
    ],
  },
};

export const mockAboutConfig = {
  hero: {
    label: "About Us",
    title: "Test About Hero",
  },
  companyProfile: {
    title: "Company Profile",
    description: "Test company profile description",
  },
  section2: {
    description: "Section 2 Content",
    image: "/images/about-section-2.png",
  },
  section3: {
    description: "Section 3 Content",
    image: "/images/about-section-3.png",
  },
};

export const mockAdvantagesConfig = {
  hero: {
    label: "Advantages",
    title: "Test Advantages Hero",
  },
  blocks: [
    {
      title: "Advantage Block 1",
      description: "Advantage Block 1 Description",
      image: "/images/advantage-1.png",
    },
    {
      title: "Advantage Block 2",
      description: "Advantage Block 2 Description",
      image: "/images/advantage-2.png",
    },
  ],
};

export const mockCoreServicesConfig = {
  hero: {
    title: "Test Core Services Hero",
    subtitle: "Test Core Services Hero Subtitle",
  },
  services: [
    {
      title: "Core Service 1",
      description: "Core Service 1 Description",
      features: ["Feature 1", "Feature 2"],
    },
  ],
};

export const mockCasesConfig = {
  hero: {
    label: "Cases",
    title: "Test Cases Hero",
  },
  cases: [
    {
      title: "Category 1",
      items: ["Case Item 1", "Case Item 2"],
    },
    {
      title: "Category 2",
      items: ["Case Item 3", "Case Item 4"],
    },
  ],
};

export const mockExpertsConfig = {
  hero: {
    label: "Experts",
    title: "Test Experts Hero",
  },
  sectionTitle: "Our Expert Team",
  experts: [
    {
      id: "expert-1",
      name: "Expert 1",
      title: "Expert 1 Title",
      bio: "Expert 1 Bio",
      image: "/images/expert-1.png",
    },
    {
      id: "expert-2",
      name: "Expert 2",
      title: "Expert 2 Title",
      bio: "Expert 2 Bio",
      image: "/images/expert-2.png",
    },
  ],
};

export const mockContactConfig = {
  hero: {
    title: "Test Contact Hero",
    subtitle: "Test Contact Hero Subtitle",
  },
  contactInfo: {
    phone: "+86 123 4567 8901",
    address: "Test Address, Beijing, China",
  },
  form: {
    title: "Contact Our Experts",
    subtitle: "Get in touch with us",
    nameLabel: "Name",
    emailLabel: "Email",
    messageLabel: "Message",
    submit: "Submit",
  },
};

export const mockGlobalConfig = {
  nav: {
    items: [
      { label: "Home", href: "/" },
      { label: "About", href: "/about" },
      { label: "Services", href: "/core-services" },
      { label: "Cases", href: "/cases" },
      { label: "Experts", href: "/experts" },
      { label: "Contact", href: "/contact" },
    ],
  },
  branding: {
    logo: "/images/logo.png",
    companyName: "Test Company",
  },
  footer: {
    copyright: "© 2026 Test Company. All rights reserved.",
    links: [
      { label: "Privacy", href: "/privacy" },
      { label: "Terms", href: "/terms" },
    ],
  },
};

export const mockConfigByPageKey = {
  home: mockHomeConfig,
  about: mockAboutConfig,
  advantages: mockAdvantagesConfig,
  "core-services": mockCoreServicesConfig,
  cases: mockCasesConfig,
  experts: mockExpertsConfig,
  contact: mockContactConfig,
  global: mockGlobalConfig,
};
