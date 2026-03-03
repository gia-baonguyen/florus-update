import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Banner } from '../components/Banner';
import { StaticBanner } from '../components/StaticBanner';
import { ProductCarousel } from '../components/ProductCarousel';
import { ProductCard } from '../components/ProductCard';
import { productsApi, recommendationsApi, categoriesApi } from '../api/products';
import { Product, Category } from '../types';
import { Sparkles, TrendingUp, Flame, Loader2, ArrowRight, ChevronRight } from 'lucide-react';

interface HomePageProps {
  onProductClick: (productId: string) => void;
  hasUserActivity: boolean;
  lastViewedProduct?: string;
  useStaticBanner?: boolean;
}

export function HomePage({ onProductClick, hasUserActivity, lastViewedProduct, useStaticBanner = false }: HomePageProps) {
  const navigate = useNavigate();
  const [products, setProducts] = useState<Product[]>([]);
  const [coldStartProducts, setColdStartProducts] = useState<Product[]>([]);
  const [warmStartProducts, setWarmStartProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        // 1. Fetch data cơ bản
        const [productsData, categoriesData] = await Promise.all([
          productsApi.getAll(1, 50),
          categoriesApi.getAll().catch(() => []),
        ]);

        setProducts(productsData.products);
        setCategories(categoriesData);

        // 2. Fetch recommendations song song với xử lý lỗi riêng biệt (Settled)
        const [coldRes, warmRes] = await Promise.allSettled([
          recommendationsApi.getColdStart(),
          recommendationsApi.getWarmStart()
        ]);

        // Xử lý Cold Start
        if (coldRes.status === 'fulfilled') {
          setColdStartProducts(coldRes.value);
        } else {
          setColdStartProducts(productsData.products.slice(0, 6));
        }

        // Xử lý Warm Start
        if (warmRes.status === 'fulfilled') {
          setWarmStartProducts(warmRes.value);
        } else {
          const shuffled = [...productsData.products].sort(() => Math.random() - 0.5);
          setWarmStartProducts(shuffled.slice(0, 6));
        }

      } catch (err) {
        console.error('Error fetching data:', err);
        setError('Unable to load data. Please try again later.');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const trendingProducts = products.filter(p => p.ai_score && p.ai_score > 80).slice(0, 8);
  const lastProduct = lastViewedProduct ? products.find(p => p.id.toString() === lastViewedProduct) : null;

  const getCategoryEmoji = (slug: string): string => {
    const emojiMap: Record<string, string> = {
      'makeup': '💄', 'skincare': '✨', 'fragrance': '🌸', 'nails': '💅',
      'hair-care': '💇', 'body-care': '🧴', 'tools': '🔧', 'sets': '🎁',
      'lips': '💋', 'eyes': '👁️', 'face': '🌟', 'suncare': '☀️', 'mens': '👔',
    };
    return emojiMap[slug] || '🎀';
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)] flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-10 h-10 text-[var(--color-primary)] animate-spin mx-auto mb-4" />
          <p className="text-gray-500 font-medium">Loading beauty world...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)] flex items-center justify-center">
        <div className="text-center p-8 bg-white rounded-2xl shadow-lg max-w-md">
          <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <span className="text-2xl">😢</span>
          </div>
          <h3 className="text-lg font-semibold text-gray-900 mb-2">Connection Error</h3>
          <p className="text-gray-500 mb-4">{error}</p>
          <button onClick={() => window.location.reload()} className="px-6 py-2 bg-[var(--color-primary)] text-white rounded-lg">
            Try Again
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      <div className="container py-6 space-y-10">
        {useStaticBanner ? <StaticBanner /> : <Banner />}

        {/* SECTION: COLD START (New Users) */}
        {!hasUserActivity && (
          <div className="space-y-6">
            <div className="bg-gradient-to-r from-[var(--color-primary)] via-[var(--color-rose-gold)] to-[var(--color-primary)] rounded-2xl p-6 text-white shadow-[var(--shadow-lg)]">
              <div className="flex items-center gap-3 mb-2">
                <Flame className="w-6 h-6" />
                <h3 className="text-white font-serif text-xl">Beauty Trends</h3>
              </div>
              <p className="text-white/90 text-sm">Discover loved products from the community. Data powered by Amazon All_Beauty.</p>
            </div>

            <div className="bg-white rounded-2xl p-6 shadow-[var(--shadow-sm)] border border-[var(--color-border)]">
              <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-2">
                  <TrendingUp className="w-5 h-5 text-[var(--color-secondary)]" />
                  <h2 className="text-2xl font-serif">Community Favorites</h2>
                </div>
           <button 
  onClick={() => navigate('/products')}
  className="px-4 py-2 rounded-full border border-[var(--color-primary)] text-[var(--color-primary)] hover:bg-[var(--color-primary)] hover:text-white transition-all duration-300 text-sm font-bold flex items-center gap-2 group"
>
  View All 
  <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
</button>
              </div>
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-4">
                {coldStartProducts.map((product) => (
                  <ProductCard key={product.id} product={product} onClick={() => onProductClick(product.id.toString())} compact />
                ))}
              </div>
            </div>
          </div>
        )}

        {/* SECTION: WARM START (Returning Users) */}
        {hasUserActivity && (
          <div className="space-y-10">
            <div className="bg-gradient-to-br from-[var(--color-secondary-light)] via-[var(--color-primary-light)] to-pink-50 rounded-2xl p-6 border border-[var(--color-secondary)]/20 shadow-[var(--shadow-md)]">
              <div className="flex items-center gap-3 mb-2">
                <div className="p-2 bg-white rounded-xl shadow-sm">
                  <Sparkles className="w-5 h-5 text-[var(--color-secondary)]" />
                </div>
                <h3 className="text-[var(--color-secondary)] font-serif text-xl">Just For You</h3>
              </div>
              <p className="text-[var(--color-text-secondary)] text-sm italic">
                {lastProduct ? (
                  <>Since you just viewed <span className="font-bold text-[var(--color-primary)]">"{lastProduct.name}"</span>, we suggest:</>
                ) : (
                  <>AI-powered recommendations based on your unique beauty profile.</>
                )}
              </p>
            </div>

            <ProductCarousel
              title="Continue Exploring"
              subtitle="Selected by AI just for you"
              products={warmStartProducts}
              onProductClick={(id) => onProductClick(id.toString())}
              recommendationLabel="AI Pick"
            />

            <div className="bg-white rounded-2xl p-6 shadow-[var(--shadow-sm)] border border-[var(--color-border)]">
              <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-2">
                  <TrendingUp className="w-6 h-6 text-[var(--color-primary)]" />
                  <h2 className="text-2xl font-serif">Trending Products</h2>
                </div>
                <button 
                  onClick={() => navigate('/products')}
                  className="flex items-center gap-1 text-[var(--color-primary)] font-bold text-sm hover:gap-2 transition-all uppercase tracking-wider"
                >
                  View All <ArrowRight className="w-4 h-4" />
                </button>
              </div>
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {trendingProducts.map((product) => (
                  <ProductCard key={product.id} product={product} onClick={() => onProductClick(product.id.toString())} compact />
                ))}
              </div>
            </div>
          </div>
        )}

        {/* SECTION: CATEGORIES */}
        <div className="bg-white rounded-2xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-sm)]">
          <div className="flex items-center justify-between mb-8">
            <h2 className="text-2xl font-serif">Product Categories</h2>
   <button 
  onClick={() => navigate('/categories')}
  className="px-4 py-2 rounded-full border border-[var(--color-primary)] text-[var(--color-primary)] hover:bg-[var(--color-primary)] hover:text-white transition-all duration-300 text-sm font-bold flex items-center gap-2 group"
>
  All Categories 
<ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" /></button>
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-4">
            {(categories.length > 0 ? categories.slice(0, 8) : [
              { id: '1', name: 'Makeup', slug: 'makeup' },
              { id: '2', name: 'Skincare', slug: 'skincare' },
              { id: '3', name: 'Fragrance', slug: 'fragrance' },
              { id: '4', name: 'Hair Care', slug: 'hair-care' },
              { id: '5', name: 'Body Care', slug: 'body-care' }
            ]).map((category) => (
              <button
                key={category.id}
                onClick={() => navigate(`/products?category=${category.id}`)}
                className="group flex flex-col items-center p-6 rounded-2xl bg-[var(--color-surface)] border-2 border-transparent hover:border-[var(--color-primary)] hover:bg-white hover:shadow-md transition-all"
              >
                <div className="text-4xl mb-3 group-hover:scale-110 transition-transform">
                  {getCategoryEmoji(category.slug)}
                </div>
                <span className="font-bold text-gray-800 group-hover:text-[var(--color-primary)]">{category.name}</span>
              </button>
            ))}
          </div>
        </div>

        {/* TECH FOOTER */}
        <div className="bg-gradient-to-br from-white to-[var(--color-surface)] rounded-2xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-sm)]">
          <h3 className="mb-6 font-serif text-lg">About Florus AI Recommendation System</h3>
          <div className="grid md:grid-cols-2 gap-6">
            <div className="p-5 bg-white rounded-xl border border-[var(--color-border)] shadow-sm flex gap-4">
              <div className="w-12 h-12 shrink-0 bg-orange-100 rounded-lg flex items-center justify-center text-orange-600">
                <Flame className="w-6 h-6" />
              </div>
              <div>
                <h4 className="font-bold text-gray-900 mb-1 tracking-tight">Cold-start Engine</h4>
                <p className="text-xs text-gray-500 leading-relaxed">Powered by <strong>Redis Cache</strong> and Popularity Scoring from Amazon's All_Beauty dataset for real-time trending.</p>
              </div>
            </div>
            <div className="p-5 bg-white rounded-xl border border-[var(--color-border)] shadow-sm flex gap-4">
              <div className="w-12 h-12 shrink-0 bg-pink-100 rounded-lg flex items-center justify-center text-pink-600">
                <Sparkles className="w-6 h-6" />
              </div>
              <div>
                <h4 className="font-bold text-gray-900 mb-1 tracking-tight">Warm-start AI Model</h4>
                <p className="text-xs text-gray-500 leading-relaxed">Personalized using <strong>Collaborative Filtering</strong> to analyze your behavior and suggest high-affinity products.</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}