import { useState, useEffect, useCallback } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { Loader2, Filter, ChevronDown, Grid, LayoutGrid, Package, Search, X } from 'lucide-react';
import { categoriesApi, productsApi } from '../api/products';
import { Category, Product } from '../types';
import { ProductCard } from '../components/ProductCard';

interface CategoryPageProps {
  onProductClick: (productId: string) => void;
}

type SortOption = 'newest' | 'price_asc' | 'price_desc' | 'popular' | 'rating';
type ViewMode = 'grid' | 'compact';

export function CategoryPage({ onProductClick }: CategoryPageProps) {
  const { categoryId } = useParams<{ categoryId: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const [categories, setCategories] = useState<Category[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [brands, setBrands] = useState<string[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<Category | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadingProducts, setLoadingProducts] = useState(false);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [sortBy, setSortBy] = useState<SortOption>('newest');
  const [viewMode, setViewMode] = useState<ViewMode>('grid');
  const [showFilters, setShowFilters] = useState(false);
  const [priceRange, setPriceRange] = useState<[number, number]>([0, 10000000]);
  const [appliedPriceRange, setAppliedPriceRange] = useState<[number, number] | null>(null);
  const [selectedBrand, setSelectedBrand] = useState<string>('');
  const [searchQuery, setSearchQuery] = useState<string>(searchParams.get('q') || '');

  useEffect(() => {
    loadCategories();
    loadBrands();
  }, []);

  const loadProducts = useCallback(async () => {
    try {
      setLoadingProducts(true);

      const data = await productsApi.getAllWithFilter({
        page,
        limit: 12,
        categoryId: categoryId ? parseInt(categoryId) : undefined,
        search: searchQuery || undefined,
        minPrice: appliedPriceRange ? appliedPriceRange[0] : undefined,
        maxPrice: appliedPriceRange ? appliedPriceRange[1] : undefined,
        brand: selectedBrand || undefined,
        sortBy,
      });

      setProducts(data.products);
      setTotalPages(data.meta?.total_pages || 1);

      if (categoryId) {
        const categoryData = await categoriesApi.getById(parseInt(categoryId));
        setSelectedCategory(categoryData);
      } else {
        setSelectedCategory(null);
      }
    } catch (err) {
      console.error('Failed to load products:', err);
    } finally {
      setLoading(false);
      setLoadingProducts(false);
    }
  }, [categoryId, page, sortBy, searchQuery, appliedPriceRange, selectedBrand]);

  useEffect(() => {
    loadProducts();
  }, [loadProducts]);

  const loadCategories = async () => {
    try {
      const data = await categoriesApi.getAll();
      setCategories(data);
    } catch (err) {
      console.error('Failed to load categories:', err);
    }
  };

  const loadBrands = async () => {
    try {
      const data = await productsApi.getBrands();
      setBrands(data);
    } catch (err) {
      console.error('Failed to load brands:', err);
    }
  };

  const handleCategorySelect = (category: Category | null) => {
    setPage(1);
    if (category) {
      window.history.pushState({}, '', `/categories/${category.id}`);
    } else {
      window.history.pushState({}, '', '/categories');
    }
  };

  const handleApplyPriceFilter = () => {
    setPage(1);
    setAppliedPriceRange([priceRange[0], priceRange[1]]);
  };

  const handleClearPriceFilter = () => {
    setPriceRange([0, 10000000]);
    setAppliedPriceRange(null);
    setPage(1);
  };

  const handleBrandSelect = (brand: string) => {
    setSelectedBrand(brand === selectedBrand ? '' : brand);
    setPage(1);
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setPage(1);
  };

  const clearSearch = () => {
    setSearchQuery('');
    setPage(1);
  };

  const clearAllFilters = () => {
    setSearchQuery('');
    setSelectedBrand('');
    setPriceRange([0, 10000000]);
    setAppliedPriceRange(null);
    setPage(1);
  };

  const hasActiveFilters = searchQuery || selectedBrand || appliedPriceRange;

  const sortOptions: { value: SortOption; label: string }[] = [
    { value: 'newest', label: 'Mới nhất' },
    { value: 'price_asc', label: 'Giá thấp đến cao' },
    { value: 'price_desc', label: 'Giá cao đến thấp' },
    { value: 'popular', label: 'Phổ biến nhất' },
    { value: 'rating', label: 'Đánh giá cao' },
  ];

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND',
    }).format(price);
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
            Trang chủ
          </a>
          <span className="text-[var(--color-text-muted)]">/</span>
          <span className="text-[var(--color-text)]">
            {selectedCategory ? selectedCategory.name : 'Tất cả sản phẩm'}
          </span>
        </nav>

        <div className="flex gap-6">
          {/* Sidebar - Categories */}
          <aside className="hidden lg:block w-64 flex-shrink-0">
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-4 sticky top-20">
              <h3 className="font-serif mb-4">Danh mục</h3>
              <ul className="space-y-1">
                <li>
                  <button
                    onClick={() => handleCategorySelect(null)}
                    className={`w-full text-left px-3 py-2 rounded-lg transition-colors ${
                      !selectedCategory
                        ? 'bg-[var(--color-primary)] text-white'
                        : 'hover:bg-[var(--color-surface)]'
                    }`}
                  >
                    Tất cả sản phẩm
                  </button>
                </li>
                {categories.map((category) => (
                  <li key={category.id}>
                    <button
                      onClick={() => handleCategorySelect(category)}
                      className={`w-full text-left px-3 py-2 rounded-lg transition-colors flex items-center gap-2 ${
                        selectedCategory?.id === category.id
                          ? 'bg-[var(--color-primary)] text-white'
                          : 'hover:bg-[var(--color-surface)]'
                      }`}
                    >
                      {category.icon && <span>{category.icon}</span>}
                      <span>{category.name}</span>
                      {category.product_count !== undefined && (
                        <span className={`ml-auto text-xs ${
                          selectedCategory?.id === category.id
                            ? 'text-white/70'
                            : 'text-[var(--color-text-muted)]'
                        }`}>
                          ({category.product_count})
                        </span>
                      )}
                    </button>
                  </li>
                ))}
              </ul>

              {/* Price Filter */}
              <div className="mt-6 pt-6 border-t border-[var(--color-border)]">
                <h4 className="font-medium mb-3">Khoảng giá</h4>
                <div className="space-y-2">
                  <div className="flex gap-2">
                    <input
                      type="number"
                      value={priceRange[0]}
                      onChange={(e) => setPriceRange([parseInt(e.target.value) || 0, priceRange[1]])}
                      placeholder="Từ"
                      className="w-full px-3 py-2 text-sm border border-[var(--color-border)] rounded-lg"
                    />
                    <span className="text-[var(--color-text-muted)]">-</span>
                    <input
                      type="number"
                      value={priceRange[1]}
                      onChange={(e) => setPriceRange([priceRange[0], parseInt(e.target.value) || 0])}
                      placeholder="Đến"
                      className="w-full px-3 py-2 text-sm border border-[var(--color-border)] rounded-lg"
                    />
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={handleApplyPriceFilter}
                      className="flex-1 py-2 bg-[var(--color-primary-light)] text-[var(--color-primary)] rounded-lg text-sm font-medium hover:bg-[var(--color-primary)] hover:text-white transition-colors"
                    >
                      Áp dụng
                    </button>
                    {appliedPriceRange && (
                      <button
                        onClick={handleClearPriceFilter}
                        className="px-3 py-2 text-[var(--color-text-muted)] hover:text-red-500 rounded-lg text-sm transition-colors"
                      >
                        <X className="w-4 h-4" />
                      </button>
                    )}
                  </div>
                  {appliedPriceRange && (
                    <p className="text-xs text-green-600">
                      Đang lọc: {formatPrice(appliedPriceRange[0])} - {formatPrice(appliedPriceRange[1])}
                    </p>
                  )}
                </div>
              </div>

              {/* Brand Filter */}
              {brands.length > 0 && (
                <div className="mt-6 pt-6 border-t border-[var(--color-border)]">
                  <h4 className="font-medium mb-3">Thương hiệu</h4>
                  <div className="space-y-1 max-h-48 overflow-y-auto">
                    {brands.map((brand) => (
                      <button
                        key={brand}
                        onClick={() => handleBrandSelect(brand)}
                        className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                          selectedBrand === brand
                            ? 'bg-[var(--color-primary)] text-white'
                            : 'hover:bg-[var(--color-surface)]'
                        }`}
                      >
                        {brand}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Clear All Filters */}
              {hasActiveFilters && (
                <div className="mt-6 pt-6 border-t border-[var(--color-border)]">
                  <button
                    onClick={clearAllFilters}
                    className="w-full py-2 text-red-500 hover:bg-red-50 rounded-lg text-sm font-medium transition-colors"
                  >
                    Xóa tất cả bộ lọc
                  </button>
                </div>
              )}
            </div>
          </aside>

          {/* Main Content */}
          <main className="flex-1">
            {/* Category Header */}
            {selectedCategory && (
              <div className="bg-white rounded-xl border border-[var(--color-border)] p-6 mb-6">
                <div className="flex items-center gap-4">
                  {selectedCategory.icon && (
                    <div className="text-4xl">{selectedCategory.icon}</div>
                  )}
                  <div>
                    <h1 className="text-2xl font-serif">{selectedCategory.name}</h1>
                    {selectedCategory.description && (
                      <p className="text-[var(--color-text-secondary)] mt-1">
                        {selectedCategory.description}
                      </p>
                    )}
                  </div>
                </div>
              </div>
            )}

            {/* Search Bar */}
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-4 mb-4">
              <form onSubmit={handleSearch} className="flex gap-2">
                <div className="flex-1 relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--color-text-muted)]" />
                  <input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Tìm kiếm sản phẩm..."
                    className="w-full pl-10 pr-10 py-2.5 border border-[var(--color-border)] rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)] focus:border-transparent"
                  />
                  {searchQuery && (
                    <button
                      type="button"
                      onClick={clearSearch}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  )}
                </div>
                <button
                  type="submit"
                  className="px-6 py-2.5 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-hover)] transition-colors"
                >
                  Tìm kiếm
                </button>
              </form>
            </div>

            {/* Active Filters */}
            {hasActiveFilters && (
              <div className="bg-white rounded-xl border border-[var(--color-border)] p-3 mb-4">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="text-sm text-[var(--color-text-secondary)]">Đang lọc:</span>
                  {searchQuery && (
                    <span className="inline-flex items-center gap-1 px-3 py-1 bg-[var(--color-primary-light)] text-[var(--color-primary)] rounded-full text-sm">
                      Tìm kiếm: "{searchQuery}"
                      <button onClick={clearSearch} className="hover:text-red-500">
                        <X className="w-3 h-3" />
                      </button>
                    </span>
                  )}
                  {selectedBrand && (
                    <span className="inline-flex items-center gap-1 px-3 py-1 bg-[var(--color-primary-light)] text-[var(--color-primary)] rounded-full text-sm">
                      Thương hiệu: {selectedBrand}
                      <button onClick={() => setSelectedBrand('')} className="hover:text-red-500">
                        <X className="w-3 h-3" />
                      </button>
                    </span>
                  )}
                  {appliedPriceRange && (
                    <span className="inline-flex items-center gap-1 px-3 py-1 bg-[var(--color-primary-light)] text-[var(--color-primary)] rounded-full text-sm">
                      Giá: {formatPrice(appliedPriceRange[0])} - {formatPrice(appliedPriceRange[1])}
                      <button onClick={handleClearPriceFilter} className="hover:text-red-500">
                        <X className="w-3 h-3" />
                      </button>
                    </span>
                  )}
                  <button
                    onClick={clearAllFilters}
                    className="text-sm text-red-500 hover:underline ml-2"
                  >
                    Xóa tất cả
                  </button>
                </div>
              </div>
            )}

            {/* Toolbar */}
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-4 mb-6">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div className="flex items-center gap-4">
                  {/* Mobile Filter Button */}
                  <button
                    onClick={() => setShowFilters(!showFilters)}
                    className="lg:hidden flex items-center gap-2 px-4 py-2 border border-[var(--color-border)] rounded-lg"
                  >
                    <Filter className="w-4 h-4" />
                    Bộ lọc
                    {hasActiveFilters && (
                      <span className="w-2 h-2 bg-[var(--color-primary)] rounded-full"></span>
                    )}
                  </button>

                  <p className="text-sm text-[var(--color-text-secondary)]">
                    Hiển thị <strong>{products.length}</strong> sản phẩm
                  </p>
                </div>

                <div className="flex items-center gap-3">
                  {/* Sort */}
                  <div className="relative">
                    <select
                      value={sortBy}
                      onChange={(e) => setSortBy(e.target.value as SortOption)}
                      className="appearance-none px-4 py-2 pr-8 border border-[var(--color-border)] rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)]"
                    >
                      {sortOptions.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                    <ChevronDown className="absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--color-text-muted)] pointer-events-none" />
                  </div>

                  {/* View Mode */}
                  <div className="hidden sm:flex items-center border border-[var(--color-border)] rounded-lg overflow-hidden">
                    <button
                      onClick={() => setViewMode('grid')}
                      className={`p-2 ${viewMode === 'grid' ? 'bg-[var(--color-primary)] text-white' : 'hover:bg-[var(--color-surface)]'}`}
                    >
                      <Grid className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => setViewMode('compact')}
                      className={`p-2 ${viewMode === 'compact' ? 'bg-[var(--color-primary)] text-white' : 'hover:bg-[var(--color-surface)]'}`}
                    >
                      <LayoutGrid className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            </div>

            {/* Products Grid */}
            {loadingProducts ? (
              <div className="flex items-center justify-center h-64">
                <Loader2 className="w-8 h-8 text-[var(--color-primary)] animate-spin" />
              </div>
            ) : products.length === 0 ? (
              <div className="bg-white rounded-xl border border-[var(--color-border)] p-12 text-center">
                <Package className="w-16 h-16 text-gray-300 mx-auto mb-4" />
                <h3 className="font-serif mb-2">Không tìm thấy sản phẩm</h3>
                <p className="text-[var(--color-text-secondary)]">
                  Thử điều chỉnh bộ lọc hoặc chọn danh mục khác
                </p>
              </div>
            ) : (
              <div className={`grid gap-4 ${
                viewMode === 'grid'
                  ? 'grid-cols-2 md:grid-cols-3 xl:grid-cols-4'
                  : 'grid-cols-2 md:grid-cols-4 xl:grid-cols-5'
              }`}>
                {products.map((product) => (
                  <ProductCard
                    key={product.id}
                    product={product}
                    onClick={() => onProductClick(product.id.toString())}
                    variant={viewMode === 'compact' ? 'compact' : 'default'}
                  />
                ))}
              </div>
            )}

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex justify-center mt-8">
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => setPage(Math.max(1, page - 1))}
                    disabled={page === 1}
                    className="px-4 py-2 border border-[var(--color-border)] rounded-lg hover:bg-[var(--color-surface)] disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Trước
                  </button>

                  {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                    let pageNum;
                    if (totalPages <= 5) {
                      pageNum = i + 1;
                    } else if (page <= 3) {
                      pageNum = i + 1;
                    } else if (page >= totalPages - 2) {
                      pageNum = totalPages - 4 + i;
                    } else {
                      pageNum = page - 2 + i;
                    }
                    return (
                      <button
                        key={pageNum}
                        onClick={() => setPage(pageNum)}
                        className={`w-10 h-10 rounded-lg ${
                          page === pageNum
                            ? 'bg-[var(--color-primary)] text-white'
                            : 'border border-[var(--color-border)] hover:bg-[var(--color-surface)]'
                        }`}
                      >
                        {pageNum}
                      </button>
                    );
                  })}

                  <button
                    onClick={() => setPage(Math.min(totalPages, page + 1))}
                    disabled={page === totalPages}
                    className="px-4 py-2 border border-[var(--color-border)] rounded-lg hover:bg-[var(--color-surface)] disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Sau
                  </button>
                </div>
              </div>
            )}
          </main>
        </div>
      </div>

      {/* Mobile Filters Modal */}
      {showFilters && (
        <div className="lg:hidden fixed inset-0 bg-black/50 z-50">
          <div className="absolute bottom-0 left-0 right-0 bg-white rounded-t-2xl max-h-[80vh] overflow-y-auto">
            <div className="p-4 border-b border-[var(--color-border)] flex items-center justify-between sticky top-0 bg-white">
              <h3 className="font-serif">Bộ lọc</h3>
              <button onClick={() => setShowFilters(false)}>
                <X className="w-6 h-6" />
              </button>
            </div>

            <div className="p-4 space-y-6">
              {/* Categories */}
              <div>
                <h4 className="font-medium mb-3">Danh mục</h4>
                <div className="space-y-2">
                  <button
                    onClick={() => {
                      handleCategorySelect(null);
                      setShowFilters(false);
                    }}
                    className={`w-full text-left px-3 py-2 rounded-lg ${
                      !selectedCategory ? 'bg-[var(--color-primary)] text-white' : 'bg-[var(--color-surface)]'
                    }`}
                  >
                    Tất cả sản phẩm
                  </button>
                  {categories.map((category) => (
                    <button
                      key={category.id}
                      onClick={() => {
                        handleCategorySelect(category);
                        setShowFilters(false);
                      }}
                      className={`w-full text-left px-3 py-2 rounded-lg flex items-center gap-2 ${
                        selectedCategory?.id === category.id
                          ? 'bg-[var(--color-primary)] text-white'
                          : 'bg-[var(--color-surface)]'
                      }`}
                    >
                      {category.icon && <span>{category.icon}</span>}
                      {category.name}
                    </button>
                  ))}
                </div>
              </div>

              {/* Price Range */}
              <div>
                <h4 className="font-medium mb-3">Khoảng giá</h4>
                <div className="space-y-2">
                  <div className="flex gap-2">
                    <input
                      type="number"
                      value={priceRange[0]}
                      onChange={(e) => setPriceRange([parseInt(e.target.value) || 0, priceRange[1]])}
                      placeholder="Từ"
                      className="w-full px-3 py-2 text-sm border border-[var(--color-border)] rounded-lg"
                    />
                    <span className="text-[var(--color-text-muted)] self-center">-</span>
                    <input
                      type="number"
                      value={priceRange[1]}
                      onChange={(e) => setPriceRange([priceRange[0], parseInt(e.target.value) || 0])}
                      placeholder="Đến"
                      className="w-full px-3 py-2 text-sm border border-[var(--color-border)] rounded-lg"
                    />
                  </div>
                  <button
                    onClick={() => {
                      handleApplyPriceFilter();
                      setShowFilters(false);
                    }}
                    className="w-full py-2 bg-[var(--color-primary-light)] text-[var(--color-primary)] rounded-lg text-sm font-medium"
                  >
                    Áp dụng giá
                  </button>
                </div>
              </div>

              {/* Brands */}
              {brands.length > 0 && (
                <div>
                  <h4 className="font-medium mb-3">Thương hiệu</h4>
                  <div className="flex flex-wrap gap-2">
                    {brands.map((brand) => (
                      <button
                        key={brand}
                        onClick={() => {
                          handleBrandSelect(brand);
                          setShowFilters(false);
                        }}
                        className={`px-3 py-1.5 rounded-full text-sm transition-colors ${
                          selectedBrand === brand
                            ? 'bg-[var(--color-primary)] text-white'
                            : 'bg-[var(--color-surface)] hover:bg-[var(--color-primary-light)]'
                        }`}
                      >
                        {brand}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Clear All */}
              {hasActiveFilters && (
                <button
                  onClick={() => {
                    clearAllFilters();
                    setShowFilters(false);
                  }}
                  className="w-full py-3 text-red-500 border border-red-200 rounded-lg font-medium"
                >
                  Xóa tất cả bộ lọc
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
