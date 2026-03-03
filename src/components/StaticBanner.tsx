import { Sparkles } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

const banner = {
  title: 'Beauty Made for You',
  subtitle: 'Discover our new collection - Up to 40% off',
  bgGradient: 'linear-gradient(135deg, rgba(255, 107, 107, 0.08) 0%, rgba(255, 107, 107, 0.03) 100%)',
};

export function StaticBanner() {
  const navigate = useNavigate();

  return (
    <div className="relative w-full h-64 md:h-80 rounded-xl overflow-hidden bg-gradient-to-br from-gray-50 to-white border border-gray-100">
      {/* Static Banner */}
      <div
        className="w-full h-full flex items-center justify-center px-8 md:px-16 relative"
        style={{ background: banner.bgGradient }}
      >
        {/* Main Content */}
        <div className="text-center relative z-10 w-full max-w-3xl">
          <h1 className="text-gray-900 mb-4">
            {banner.title}
          </h1>
          
          <p className="text-gray-600 mb-8 text-lg">
            {banner.subtitle}
          </p>
          
          <button
            onClick={() => navigate('/products')}
            className="px-8 py-3.5 bg-[var(--color-primary)] text-white rounded-lg hover:opacity-90 transition-all inline-block"
          >
            Shop Now
          </button>
        </div>
      </div>

      {/* Decorative element */}
      <div className="absolute top-6 right-6 text-[var(--color-primary)] opacity-20">
        <Sparkles className="w-8 h-8" />
      </div>
    </div>
  );
}