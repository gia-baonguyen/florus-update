import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Loader2, Package, ArrowRight } from 'lucide-react';
import { categoriesApi } from '../api/products';
import { Category } from '../types';

export function CategoriesListPage() {
  const navigate = useNavigate();
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadCategories();
  }, []);

  const loadCategories = async () => {
    try {
      const data = await categoriesApi.getAll();
      setCategories(data);
    } catch (err) {
      console.error('Failed to load categories:', err);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)] flex items-center justify-center">
        <Loader2 className="w-10 h-10 text-[var(--color-primary)] animate-spin" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      <div className="container py-6">
        {/* Breadcrumb */}
        <nav className="flex items-center gap-2 text-sm mb-6">
          <a href="/" className="text-[var(--color-text-secondary)] hover:text-[var(--color-primary)]">
            Home
          </a>
          <span className="text-[var(--color-text-muted)]">/</span>
          <span className="text-[var(--color-text)]">Categories</span>
        </nav>

        {/* Page Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-serif mb-2">Shop by Category</h1>
          <p className="text-[var(--color-text-secondary)]">
            Browse our collection by category to find exactly what you're looking for
          </p>
        </div>

        {/* View All Products Button */}
        <div className="mb-6">
          <button
            onClick={() => navigate('/products')}
            className="inline-flex items-center gap-2 px-6 py-3 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-dark)] transition-colors"
          >
            View All Products
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>

        {/* Categories Grid */}
        {categories.length === 0 ? (
          <div className="bg-white rounded-xl border border-[var(--color-border)] p-12 text-center">
            <Package className="w-16 h-16 text-gray-300 mx-auto mb-4" />
            <h3 className="font-serif mb-2">No Categories Found</h3>
            <p className="text-[var(--color-text-secondary)]">
              Categories will appear here once added
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {categories.map((category) => (
              <button
                key={category.id}
                onClick={() => navigate(`/products?category=${category.id}`)}
                className="group bg-white rounded-xl border border-[var(--color-border)] p-6 text-left hover:shadow-lg hover:border-[var(--color-primary)] transition-all"
              >
                {/* Category Icon/Image */}
                <div className="w-16 h-16 bg-gradient-to-br from-[var(--color-primary-light)] to-pink-50 rounded-xl flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                  {category.icon ? (
                    <span className="text-3xl">{category.icon}</span>
                  ) : (
                    <Package className="w-8 h-8 text-[var(--color-primary)]" />
                  )}
                </div>

                {/* Category Info */}
                <h3 className="font-serif text-lg mb-1 group-hover:text-[var(--color-primary)] transition-colors">
                  {category.name}
                </h3>

                {category.description && (
                  <p className="text-sm text-[var(--color-text-secondary)] line-clamp-2 mb-3">
                    {category.description}
                  </p>
                )}

                {/* Product Count */}
                {category.product_count !== undefined && (
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-[var(--color-text-muted)]">
                      {category.product_count} products
                    </span>
                    <ArrowRight className="w-4 h-4 text-[var(--color-primary)] opacity-0 group-hover:opacity-100 transition-opacity" />
                  </div>
                )}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
