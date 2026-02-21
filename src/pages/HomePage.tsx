import { useState, useEffect } from 'react';
import { Banner } from '../components/Banner';
import { StaticBanner } from '../components/StaticBanner';
import { ProductCarousel } from '../components/ProductCarousel';
import { ProductCard } from '../components/ProductCard';
import { productsApi, recommendationsApi, categoriesApi } from '../api/products';
import { Product, Category } from '../types';
import { Sparkles, TrendingUp, Flame, Loader2 } from 'lucide-react';

interface HomePageProps {
  onProductClick: (productId: string) => void;
  hasUserActivity: boolean;
  lastViewedProduct?: string;
  useStaticBanner?: boolean;
}

export function HomePage({ onProductClick, hasUserActivity, lastViewedProduct, useStaticBanner = false }: HomePageProps) {
  const [products, setProducts] = useState<Product[]>([]);
  const [coldStartProducts, setColdStartProducts] = useState<Product[]>([]);
  const [warmStartProducts, setWarmStartProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Fetch products and recommendations
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        // Fetch all data in parallel
        const [productsData, categoriesData] = await Promise.all([
          productsApi.getAll(1, 50),
          categoriesApi.getAll().catch(() => []),
        ]);

        setProducts(productsData.products);
        setCategories(categoriesData);

        // Fetch recommendations
        try {
          const coldStart = await recommendationsApi.getColdStart();
          setColdStartProducts(coldStart);
        } catch {
          // Fallback to first 6 products if recommendations API fails
          setColdStartProducts(productsData.products.slice(0, 6));
        }

        try {
          const warmStart = await recommendationsApi.getWarmStart();
          setWarmStartProducts(warmStart);
        } catch {
          // Fallback to random products if recommendations API fails
          const shuffled = [...productsData.products].sort(() => Math.random() - 0.5);
          setWarmStartProducts(shuffled.slice(0, 6));
        }

      } catch (err) {
        console.error('Error fetching data:', err);
        setError('Không thể tải dữ liệu. Vui lòng thử lại sau.');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  // Trending products for grid display
  const trendingProducts = products.filter(p => p.ai_score && p.ai_score > 80).slice(0, 8);

  // Get last viewed product name for personalization message
  const lastProduct = lastViewedProduct ? products.find(p => p.id.toString() === lastViewedProduct) : null;

  // Helper function to get emoji for category
  const getCategoryEmoji = (slug: string): string => {
    const emojiMap: Record<string, string> = {
      'makeup': '💄',
      'skincare': '✨',
      'fragrance': '🌸',
      'nails': '💅',
      'hair-care': '💇',
      'body-care': '🧴',
      'tools': '🔧',
      'sets': '🎁',
      'lips': '💋',
      'eyes': '👁️',
      'face': '🌟',
      'suncare': '☀️',
      'mens': '👔',
    };
    return emojiMap[slug] || '🎀';
  };

  // Loading state
  if (loading) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)] flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-10 h-10 text-[var(--color-primary)] animate-spin mx-auto mb-4" />
          <p className="text-gray-500">Đang tải sản phẩm...</p>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)] flex items-center justify-center">
        <div className="text-center p-8 bg-white rounded-2xl shadow-lg max-w-md">
          <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <span className="text-2xl">😢</span>
          </div>
          <h3 className="text-lg font-semibold text-gray-900 mb-2">Lỗi kết nối</h3>
          <p className="text-gray-500 mb-4">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-6 py-2 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-dark)] transition-colors"
          >
            Thử lại
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      <div className="container py-6 space-y-6">
        {/* Hero Banner Carousel */}
        {/* Use StaticBanner for easy Figma copy, or animated Banner for production */}
        {useStaticBanner ? <StaticBanner /> : <Banner />}

        {/* Cold Start Section - For new users */}
        {!hasUserActivity && (
          <div className="space-y-5">
            {/* Info Banner */}
            <div className="bg-gradient-to-r from-[var(--color-primary)] via-[var(--color-rose-gold)] to-[var(--color-primary)] rounded-2xl p-6 text-white shadow-[var(--shadow-lg)]">
              <div className="flex items-center gap-3 mb-2">
                <Flame className="w-6 h-6" />
                <h3 className="text-white font-serif">Xu hướng làm đẹp</h3>
              </div>
              <p className="text-white/95 text-sm leading-relaxed">
                Khám phá những sản phẩm làm đẹp được yêu thích nhất từ cộng đồng beauty lovers.
                Dữ liệu từ Amazon All_Beauty dataset.
              </p>
            </div>

            {/* Popular Products Grid */}
            <div className="bg-white rounded-2xl p-6 shadow-[var(--shadow-sm)] border border-[var(--color-border)]">
              <div className="flex items-center gap-2 mb-5">
                <TrendingUp className="w-5 h-5 text-[var(--color-secondary)]" />
                <h2 className="font-serif">Cộng đồng đang yêu thích</h2>
              </div>
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-4">
                {coldStartProducts.map((product) => (
                  <ProductCard
                    key={product.id}
                    product={product}
                    onClick={() => onProductClick(product.id.toString())}
                    compact
                  />
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Warm Start Section - For users with activity */}
        {hasUserActivity && (
          <>
            {/* Personalized Recommendations Banner */}
            <div className="bg-gradient-to-br from-[var(--color-secondary-light)] via-[var(--color-primary-light)] to-pink-50 rounded-2xl p-6 border border-[var(--color-secondary)]/20 shadow-[var(--shadow-md)]">
              <div className="flex items-center gap-3 mb-2">
                <div className="p-2 bg-white rounded-xl shadow-sm">
                  <Sparkles className="w-5 h-5 text-[var(--color-secondary)]" />
                </div>
                <h3 className="text-[var(--color-secondary)] font-serif">Dành riêng cho bạn</h3>
              </div>
              <p className="text-[var(--color-text-secondary)] text-sm leading-relaxed italic">
                {lastProduct ? (
                  <>Vì bạn vừa xem <span className="font-medium text-[var(--color-primary)]">"{lastProduct.name}"</span>,
                  chúng tôi gợi ý những sản phẩm phù hợp với sở thích của bạn.</>
                ) : (
                  <>Được gợi ý bởi AI dựa trên sở thích và hành vi duyệt sản phẩm của bạn.</>
                )}
              </p>
            </div>

            {/* AI Recommendations Carousel */}
            <ProductCarousel
              title="Tiếp tục khám phá"
              subtitle="Sản phẩm được AI chọn riêng cho bạn"
              products={warmStartProducts}
              onProductClick={(id) => onProductClick(id.toString())}
              recommendationLabel="AI Pick"
            />

            {/* Trending Products */}
            <div className="bg-white rounded-2xl p-6 shadow-[var(--shadow-sm)] border border-[var(--color-border)]">
              <div className="flex items-center gap-2 mb-5">
                <TrendingUp className="w-6 h-6 text-[var(--color-primary)]" />
                <h2 className="font-serif">Sản phẩm thịnh hành</h2>
              </div>
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {trendingProducts.map((product) => (
                  <ProductCard
                    key={product.id}
                    product={product}
                    onClick={() => onProductClick(product.id.toString())}
                    compact
                  />
                ))}
              </div>
            </div>
          </>
        )}

        {/* Categories Section */}
        <div className="bg-white rounded-2xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-sm)]">
          <h2 className="mb-5 font-serif">Danh mục sản phẩm</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {categories.length > 0 ? (
              categories.slice(0, 8).map((category) => (
                <button
                  key={category.id}
                  className="p-5 rounded-xl border-2 border-[var(--color-border)] hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-all group"
                >
                  <div className="text-3xl mb-2">
                    {getCategoryEmoji(category.slug)}
                  </div>
                  <p className="text-center group-hover:text-[var(--color-primary)] transition-colors">{category.name}</p>
                </button>
              ))
            ) : (
              // Fallback categories
              [
                { name: 'Makeup', emoji: '💄', slug: 'makeup' },
                { name: 'Skincare', emoji: '✨', slug: 'skincare' },
                { name: 'Fragrance', emoji: '🌸', slug: 'fragrance' },
                { name: 'Hair Care', emoji: '💇', slug: 'hair-care' },
              ].map((category) => (
                <button
                  key={category.slug}
                  className="p-5 rounded-xl border-2 border-[var(--color-border)] hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-all group"
                >
                  <div className="text-3xl mb-2">{category.emoji}</div>
                  <p className="text-center group-hover:text-[var(--color-primary)] transition-colors">{category.name}</p>
                </button>
              ))
            )}
          </div>
        </div>

        {/* Tech Info Footer - Showcasing the ML/Data Engineering */}
        <div className="bg-gradient-to-br from-white to-[var(--color-surface)] rounded-2xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-sm)]">
          <h3 className="mb-4 font-serif">Về hệ thống gợi ý Florus AI</h3>
          <div className="grid md:grid-cols-2 gap-4">
            <div className="p-5 bg-white rounded-xl border border-[var(--color-border)] shadow-sm">
              <div className="flex items-center gap-2 mb-3">
                <div className="w-8 h-8 bg-[var(--color-primary-light)] rounded-lg flex items-center justify-center">
                  <Flame className="w-4 h-4 text-[var(--color-primary)]" />
                </div>
                <h4 className="text-[var(--color-text-primary)]">Cold-start Engine</h4>
              </div>
              <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
                Với người dùng mới, chúng tôi sử dụng <strong>Popular Engine</strong> kết hợp 
                <strong> Redis cache</strong> để hiển thị sản phẩm thịnh hành từ dataset 
                Amazon All_Beauty theo thời gian thực.
              </p>
            </div>
            <div className="p-5 bg-white rounded-xl border border-[var(--color-border)] shadow-sm">
              <div className="flex items-center gap-2 mb-3">
                <div className="w-8 h-8 bg-[var(--color-secondary-light)] rounded-lg flex items-center justify-center">
                  <Sparkles className="w-4 h-4 text-[var(--color-secondary)]" />
                </div>
                <h4 className="text-[var(--color-text-primary)]">Warm-start AI Model</h4>
              </div>
              <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
                Dựa trên lịch sử xem và tương tác, <strong>AI model</strong> của chúng tôi phân tích 
                patterns và đưa ra gợi ý cá nhân hóa sử dụng <strong>Collaborative Filtering</strong>.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}