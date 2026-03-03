import { useEffect, useState } from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

const banners = [
  {
    id: 1,
    title: 'Beauty Made for You',
    subtitle: 'Discover our new collection - Up to 40% off',
    bgGradient: 'linear-gradient(135deg, rgba(255, 107, 107, 0.08) 0%, rgba(255, 107, 107, 0.03) 100%)',
  },
  {
    id: 2,
    title: 'Radiant Skin Every Day',
    subtitle: 'Vitamin C Serum - Special Discount',
    bgGradient: 'linear-gradient(135deg, rgba(31, 171, 137, 0.08) 0%, rgba(31, 171, 137, 0.03) 100%)',
  },
  {
    id: 3,
    title: 'Perfect Makeup',
    subtitle: 'Makeup Collection 2024 - Buy 2 Get 1 Free',
    bgGradient: 'linear-gradient(135deg, rgba(255, 107, 107, 0.08) 0%, rgba(31, 171, 137, 0.03) 100%)',
  },
];

export function Banner() {
  const [currentIndex, setCurrentIndex] = useState(0);
  const navigate = useNavigate();
  const currentBanner = banners[currentIndex];

  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentIndex((prev) => (prev + 1) % banners.length);
    }, 5000);

    return () => clearInterval(timer);
  }, []);

  const goToPrevious = () => {
    setCurrentIndex((prev) => (prev - 1 + banners.length) % banners.length);
  };

  const goToNext = () => {
    setCurrentIndex((prev) => (prev + 1) % banners.length);
  };

  return (
    <div className="relative w-full h-64 md:h-80 rounded-xl overflow-hidden group bg-gradient-to-br from-gray-50 to-white border border-gray-100">
      {/* Current Banner */}
      <div
        className="w-full h-full flex items-center justify-center px-8 md:px-16 relative transition-all duration-700"
        style={{ background: currentBanner.bgGradient }}
      >
        {/* Main Content */}
        <div className="text-center relative z-10 w-full max-w-3xl">
          <h1 className="text-gray-900 mb-4">
            {currentBanner.title}
          </h1>
          
          <p className="text-gray-600 mb-8 text-lg">
            {currentBanner.subtitle}
          </p>
          
          <button
            onClick={() => navigate('/products')}
            className="px-8 py-3.5 bg-[var(--color-primary)] text-white rounded-lg hover:opacity-90 transition-all inline-block"
          >
            Shop Now
          </button>
        </div>
      </div>

      {/* Navigation Buttons */}
      <button
        onClick={goToPrevious}
        className="absolute left-4 top-1/2 -translate-y-1/2 p-2 rounded-lg bg-white/80 hover:bg-white backdrop-blur-sm opacity-0 group-hover:opacity-100 transition-opacity border border-gray-200"
        aria-label="Previous banner"
      >
        <ChevronLeft className="w-5 h-5 text-gray-700" />
      </button>
      <button
        onClick={goToNext}
        className="absolute right-4 top-1/2 -translate-y-1/2 p-2 rounded-lg bg-white/80 hover:bg-white backdrop-blur-sm opacity-0 group-hover:opacity-100 transition-opacity border border-gray-200"
        aria-label="Next banner"
      >
        <ChevronRight className="w-5 h-5 text-gray-700" />
      </button>

      {/* Indicators */}
      <div className="absolute bottom-6 left-1/2 -translate-x-1/2 flex gap-2">
        {banners.map((_, index) => (
          <button
            key={index}
            onClick={() => setCurrentIndex(index)}
            className={`h-1.5 rounded-full transition-all ${
              index === currentIndex
                ? 'bg-[var(--color-primary)] w-8'
                : 'bg-gray-300 hover:bg-gray-400 w-1.5'
            }`}
            aria-label={`Go to banner ${index + 1}`}
          />
        ))}
      </div>
    </div>
  );
}