import { Banner } from '../components/Banner';
import { StaticBanner } from '../components/StaticBanner';
import { ProductCarousel } from '../components/ProductCarousel';
import { ProductCard } from '../components/ProductCard';
import { products, coldStartRecommendations, warmStartRecommendations } from '../data/products';
import { Sparkles, TrendingUp, Flame } from 'lucide-react';

interface HomePageProps {
  onProductClick: (productId: string) => void;
  hasUserActivity: boolean;
  lastViewedProduct?: string;
  useStaticBanner?: boolean; // NEW: Toggle for Figma-friendly static banner
}

export function HomePage({ onProductClick, hasUserActivity, lastViewedProduct, useStaticBanner = false }: HomePageProps) {
  // Cold Start - Popular products for new users
  const coldStartProducts = coldStartRecommendations
    .map(id => products.find(p => p.id === id))
    .filter(Boolean) as typeof products;

  // Warm Start - Personalized recommendations for users with activity
  const warmStartProducts = warmStartRecommendations
    .map(id => products.find(p => p.id === id))
    .filter(Boolean) as typeof products;

  // Trending products for grid display
  const trendingProducts = products.filter(p => p.isTrending || p.isHot);

  // Get last viewed product name for personalization message
  const lastProduct = lastViewedProduct ? products.find(p => p.id === lastViewedProduct) : null;

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
                    onClick={() => onProductClick(product.id)}
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
              onProductClick={onProductClick}
              recommendationLabel="AI Pick ✨"
            />

            {/* Trending Products */}
            <div className="bg-white rounded-2xl p-6 shadow-[var(--shadow-sm)] border border-[var(--color-border)]">
              <div className="flex items-center gap-2 mb-5">
                <TrendingUp className="w-6 h-6 text-[var(--color-primary)]" />
                <h2 className="font-serif">Sản phẩm thịnh hành</h2>
              </div>
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {trendingProducts.slice(0, 8).map((product) => (
                  <ProductCard
                    key={product.id}
                    product={product}
                    onClick={() => onProductClick(product.id)}
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
            {[
              { name: 'Makeup', emoji: '💄' },
              { name: 'Skincare', emoji: '✨' },
              { name: 'Fragrance', emoji: '🌸' },
              { name: 'Nails', emoji: '💅' },
            ].map((category) => (
              <button
                key={category.name}
                className="p-5 rounded-xl border-2 border-[var(--color-border)] hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-all group"
              >
                <div className="text-3xl mb-2">{category.emoji}</div>
                <p className="text-center group-hover:text-[var(--color-primary)] transition-colors">{category.name}</p>
              </button>
            ))}
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