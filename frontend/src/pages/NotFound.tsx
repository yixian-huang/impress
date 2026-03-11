export default function NotFoundPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center px-4">
        <h1 className="text-9xl font-bold text-gray-200">404</h1>
        <h2 className="text-2xl font-semibold text-gray-700 mt-4">
          {t("notFound.title", "Page Not Found")}
        </h2>
        <p className="text-gray-500 mt-2 max-w-md mx-auto">
          {t("notFound.description", "The page you are looking for might have been removed or is temporarily unavailable.")}
        </p>
        <div className="mt-6 flex gap-3 justify-center">
          <button
            onClick={() => navigate(-1)}
            className="px-4 py-2 text-sm text-gray-600 border rounded-lg hover:bg-gray-100"
          >
            {t("notFound.goBack", "Go Back")}
          </button>
          <Link to="/" className="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700">
            {t("notFound.home", "Home")}
          </Link>
        </div>
      </div>
    </div>
  );
}
