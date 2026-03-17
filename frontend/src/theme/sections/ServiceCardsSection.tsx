import type { SectionProps } from "../types";

export interface ServiceCard {
  title?: string;
  description?: string;
  image?: string;
  link?: string;
}

export interface ServiceCardsSectionData {
  title?: string;
  services?: ServiceCard[];
  items?: ServiceCard[];
}

export default function ServiceCardsSection({ data }: SectionProps<ServiceCardsSectionData>) {
  const { title, items, services } = data;
  const serviceList = items || services;

  return (
    <div className="max-w-layout w-full h-full mx-auto px-4 md:px-content xl:px-8">
      {title && (
        <div className="flex items-center mb-8 sm:mb-12">
          <div className="w-5 h-5 sm:w-[26px] sm:h-[26px] bg-accent mr-2 sm:mr-3 flex-shrink-0 rounded-sm" />
          <h2 className="text-2xl sm:text-3xl md:text-4xl font-bold text-primary truncate min-w-0">
            {title}
          </h2>
          <span className="ml-1 sm:ml-2 text-xl sm:text-2xl text-accent flex-shrink-0 cursor-pointer">
            &gt;
          </span>
        </div>
      )}
      {serviceList && serviceList.length > 0 && (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 sm:gap-8 xl:gap-10">
          {serviceList.map((service, index) => (
            <div
              key={index}
              className="flex flex-col rounded-lg overflow-hidden bg-white shadow-sm border border-gray-100 hover:shadow-md transition-shadow"
            >
              <div className="w-full h-[160px] sm:h-[180px] flex-shrink-0 overflow-hidden bg-surface-alt">
                <img
                  src={service.image || `/images/service-${index + 1}.png`}
                  alt={service.title || `Service ${index + 1}`}
                  className="w-full h-full object-cover"
                />
              </div>
              <div className="flex flex-col flex-1 p-4 sm:p-5">
                {service.title && (
                  <h3 className="text-base font-bold text-primary mb-2">
                    {service.title}
                  </h3>
                )}
                {service.description && (
                  <p className="text-sm text-on-surface-muted leading-relaxed mb-3 flex-1">
                    {service.description}
                  </p>
                )}
                {service.link && (
                  <a
                    href="#"
                    className="text-sm font-bold text-primary hover:text-accent transition-colors cursor-pointer"
                  >
                    {service.link}
                  </a>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
