import { useState, useEffect, useCallback } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { Loader2, Filter, ChevronDown, Grid, LayoutGrid, Package, Search, X, Star, Percent, CheckCircle } from 'lucide-react';
import { categoriesApi, productsApi } from '../api/products';
import { Category, Product } from '../types';
import { useSearchTracking } from '../hooks/useEventTracking';
import { ProductCard } from '../components/ProductCard';

interface ProductsPageProps {
  onProductClick: (productId: string) => void;
}

type SortOption = 'newest' | 'price_asc' | 'price_desc' | 'popular' | 'rating';
type ViewMode = 'grid' | 'compact';

export function ProductsPage({ onProductClick }: ProductsPageProps) {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();

  // Get initial values from URL
  const initialCategoryId = searchParams.get('category');
  const initialSearch = searchParams.get('q') || '';
  const initialBrand = searchParams.get('brand') || '';
  const initialInStock = searchParams.get('in_stock') === 'true';
  const initialOnSale = searchParams.get('on_sale') === 'true';
  const initialMinRating = parseInt(searchParams.get('min_rating') || '0');

  const [categories, setCategories] = useState<Category[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [brands, setBrands] = useState<string[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<Category | null>(null);
  const [selectedCategoryId, setSelectedCategoryId] = useState<number | null>(
    initialCategoryId ? parseInt(initialCategoryId) : null
  );
  const [loading, setLoading] = useState(true);
  const [loadingProducts, setLoadingProducts] = useState(false);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalProducts, setTotalProducts] = useState(0);
  const [sortBy, setSortBy] = useState<SortOption>('newest');
  const [viewMode, setViewMode] = useState<ViewMode>('grid');
  const [showFilters, setShowFilters] = useState(false);

  // Filter states
  const [priceRange, setPriceRange] = useState<[number, number]>([0, 10000000]);
  const [appliedPriceRange, setAppliedPriceRange] = useState<[number, number] | null>(null);
  const [selectedBrand, setSelectedBrand] = useState<string>(initialBrand);
  const [searchQuery, setSearchQuery] = useState<string>(initialSearch);
  const [inStockOnly, setInStockOnly] = useState<boolean>(initialInStock);
  const [onSaleOnly, setOnSaleOnly] = useState<boolean>(initialOnSale);
  const [minRating, setMinRating] = useState<number>(initialMinRating);
  const trackSearch = useSearchTracking();

  useEffect(() => {
    loadCategories();
    loadBrands();
  }, []);

  // Update URL when filters change
  const updateURL = useCallback(() => {
    const params = new URLSearchParams();
    if (selectedCategoryId) params.set('category', selectedCategoryId.toString());
    if (searchQuery) params.set('q', searchQuery);
    if (selectedBrand) params.set('brand', selectedBrand);
    if (inStockOnly) params.set('in_stock', 'true');
    if (onSaleOnly) params.set('on_sale', 'true');
    if (minRating > 0) params.set('min_rating', minRating.toString());

    const newSearch = params.toString();
    if (newSearch !== searchParams.toString()) {
      setSearchParams(params, { replace: true });
    }
  }, [selectedCategoryId, searchQuery, selectedBrand, inStockOnly, onSaleOnly, minRating, searchParams, setSearchParams]);

  useEffect(() => {
    updateURL();
  }, [updateURL]);

  const loadProducts = useCallback(async () => {
    try {
      setLoadingProducts(true);

      const data = await productsApi.getAllWithFilter({
        page,
        limit: 12,
        categoryId: selectedCategoryId || undefined,
        search: searchQuery || undefined,
        minPrice: appliedPriceRange ? appliedPriceRange[0] : undefined,
        maxPrice: appliedPriceRange ? appliedPriceRange[1] : undefined,
        brand: selectedBrand || undefined,
        sortBy,
        inStock: inStockOnly || undefined,
        onSale: onSaleOnly || undefined,
        minRating: minRating > 0 ? minRating : undefined,
      });

      setProducts(data.products || []);
      setTotalPages(data.meta?.total_pages || 1);
      setTotalProducts(data.meta?.total || 0);

      if (selectedCategoryId) {
        const categoryData = await categoriesApi.getById(selectedCategoryId);
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
  }, [selectedCategoryId, page, sortBy, searchQuery, appliedPriceRange, selectedBrand, inStockOnly, onSaleOnly, minRating]);

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

  const handleCategorySelect = (categoryId: number | null) => {
    setSelectedCategoryId(categoryId);
    setPage(1);
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
    if (searchQuery.trim()) {
      trackSearch(searchQuery);
    }
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
    setSelectedCategoryId(null);
    setInStockOnly(false);
    setOnSaleOnly(false);
    setMinRating(0);
    setPage(1);
  };

  const hasActiveFilters = searchQuery || selectedBrand || appliedPriceRange || selectedCategoryId || inStockOnly || onSaleOnly || minRating > 0;

  const sortOptions: { value: SortOption; label: string }[] = [
    { value: 'newest', label: 'Newest' },
    { value: 'price_asc', label: 'Price: Low to High' },
    { value: 'price_desc', label: 'Price: High to Low' },
    { value: 'popular', label: 'Most Popular' },
    { value: 'rating', label: 'Highest Rated' },
  ];

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price / 1000);
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
          <button
            onClick={() => navigate('/categories')}
            className="text-[var(--color-text-secondary)] hover:text-[var(--color-primary)]"
          >
            Categories
          </button>
          <span className="text-[var(--color-text-muted)]">/</span>
          <span className="text-[var(--color-text)]">
            {selectedCategory ? selectedCategory.name : 'All Products'}
          </span>
        </nav>

        <div className="flex gap-6">
          {/* Sidebar - Filters */}
          <aside className="hidden lg:block w-72 flex-shrink-0">
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-4 sticky top-20">
              <div className="flex items-center justify-between mb-4">
                <h3 className="font-serif">Filters</h3>
                {hasActiveFilters && (
                  <button
                    onClick={clearAllFilters}
                    className="text-xs text-red-500 hover:underline"
                  >
                    Clear All
                  </button>
                )}
              </div>

              {/* Categories */}
              <div className="mb-6">
                <h4 className="font-medium mb-3 text-sm">Categories</h4>
                <ul className="space-y-1">
                  <li>
                    <button
                      onClick={() => handleCategorySelect(null)}
                      className={`w-full text-left px-3 py-2 rounded-lg transition-colors text-sm ${
                        !selectedCategoryId
                          ? 'bg-[var(--color-primary)] text-white'
                          : 'hover:bg-[var(--color-surface)]'
                      }`}
                    >
                      All Products
                    </button>
                  </li>
                  {categories.map((category) => (
                    <li key={category.id}>
                      <button
                        onClick={() => handleCategorySelect(category.id)}
                        className={`w-full text-left px-3 py-2 rounded-lg transition-colors flex items-center gap-2 text-sm ${
                          selectedCategoryId === category.id
                            ? 'bg-[var(--color-primary)] text-white'
                            : 'hover:bg-[var(--color-surface)]'
                        }`}
                      >
                        {category.icon && <span>{category.icon}</span>}
                        <span className="flex-1">{category.name}</span>
                        {category.product_count !== undefined && (
                          <span className={`text-xs ${
                            selectedCategoryId === category.id
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
              </div>

              {/* Quick Filters */}
              <div className="mb-6 pt-4 border-t border-[var(--color-border)]">
                <h4 className="font-medium mb-3 text-sm">Quick Filters</h4>
                <div className="space-y-2">
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={inStockOnly}
                      onChange={(e) => {
                        setInStockOnly(e.target.checked);
                        setPage(1);
                      }}
                      className="w-4 h-4 rounded border-gray-300 text-[var(--color-primary)] focus:ring-[var(--color-primary)]"
                    />
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    <span className="text-sm">In Stock Only</span>
                  </label>
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={onSaleOnly}
                      onChange={(e) => {
                        setOnSaleOnly(e.target.checked);
                        setPage(1);
                      }}
                      className="w-4 h-4 rounded border-gray-300 text-[var(--color-primary)] focus:ring-[var(--color-primary)]"
                    />
                    <Percent className="w-4 h-4 text-red-500" />
                    <span className="text-sm">On Sale</span>
                  </label>
                </div>
              </div>

              {/* Rating Filter */}
              <div className="mb-6 pt-4 border-t border-[var(--color-border)]">
                <h4 className="font-medium mb-3 text-sm">Minimum Rating</h4>
                <div className="space-y-1">
                  {[4, 3, 2, 1, 0].map((rating) => (
                    <button
                      key={rating}
                      onClick={() => {
                        setMinRating(rating);
                        setPage(1);
                      }}
                      className={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
                        minRating === rating
                          ? 'bg-[var(--color-primary)] text-white'
                          : 'hover:bg-[var(--color-surface)]'
                      }`}
                    >
                      {rating > 0 ? (
                        <>
                          <div className="flex">
                            {[...Array(5)].map((_, i) => (
                              <Star
                                key={i}
                                className={`w-4 h-4 ${
                                  i < rating
                                    ? minRating === rating
                                      ? 'text-white fill-current'
                                      : 'text-yellow-400 fill-current'
                                    : minRating === rating
                                      ? 'text-white/30'
                                      : 'text-gray-300'
                                }`}
                              />
                            ))}
                          </div>
                          <span>& Up</span>
                        </>
                      ) : (
                        <span>All Ratings</span>
                      )}
                    </button>
                  ))}
                </div>
              </div>

              {/* Price Filter */}
              <div className="mb-6 pt-4 border-t border-[var(--color-border)]">
                <h4 className="font-medium mb-3 text-sm">Price Range</h4>
                <div className="space-y-2">
                  <div className="flex gap-2">
                    <input
                      type="number"
                      value={priceRange[0]}
                      onChange={(e) => setPriceRange([parseInt(e.target.value) || 0, priceRange[1]])}
                      placeholder="Min"
                      className="w-full px-3 py-2 text-sm border border-[var(--color-border)] rounded-lg"
                    />
                    <span className="text-[var(--color-text-muted)] self-center">-</span>
                    <input
                      type="number"
                      value={priceRange[1]}
                      onChange={(e) => setPriceRange([priceRange[0], parseInt(e.target.value) || 0])}
                      placeholder="Max"
                      className="w-full px-3 py-2 text-sm border border-[var(--color-border)] rounded-lg"
                    />
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={handleApplyPriceFilter}
                      className="flex-1 py-2 bg-[var(--color-primary-light)] text-[var(--color-primary)] rounded-lg text-sm font-medium hover:bg-[var(--color-primary)] hover:text-white transition-colors"
                    >
                      Apply
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
                      {formatPrice(appliedPriceRange[0])} - {formatPrice(appliedPriceRange[1])}
                    </p>
                  )}
                </div>
              </div>

              {/* Brand Filter */}
              {brands.length > 0 && (
                <div className="pt-4 border-t border-[var(--color-border)]">
                  <h4 className="font-medium mb-3 text-sm">Brands</h4>
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
            </div>
          </aside>

          {/* Main Content */}
          <main className="flex-1">
            {/* Page Header */}
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-6 mb-6">
              <div className="flex items-center gap-4">
                {selectedCategory?.icon && (
                  <div className="text-4xl">{selectedCategory.icon}</div>
                )}
                <div>
                  <h1 className="text-2xl font-serif">
                    {selectedCategory ? selectedCategory.name : 'All Products'}
                  </h1>
                  {selectedCategory?.description ? (
                    <p className="text-[var(--color-text-secondary)] mt-1">
                      {selectedCategory.description}
                    </p>
                  ) : (
                    <p className="text-[var(--color-text-secondary)] mt-1">
                      Browse our complete collection of beauty products
                    </p>
                  )}
                </div>
              </div>
            </div>

            {/* Search Bar */}
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-4 mb-4">
              <form onSubmit={handleSearch} className="flex gap-2">
                <div className="flex-1 relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--color-text-muted)]" />
                  <input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search products by name, brand, or ingredients..."
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
                  className="px-6 py-2.5 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-dark)] transition-colors"
                >
                  Search
                </button>
              </form>
            </div>

            {/* Active Filters Tags */}
            {hasActiveFilters && (
              <div className="bg-white rounded-xl border border-[var(--color-border)] p-3 mb-4">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="text-sm text-[var(--color-text-secondary)]">Active filters:</span>

                  {selectedCategoryId && selectedCategory && (
                    <span className="inline-flex items-center gap-1 px-3 py-1 bg-blue-100 text-blue-700 rounded-full text-sm">
                      Category: {selectedCategory.name}
                      <button onClick={() => handleCategorySelect(null)} className="hover:text-red-500">
                        <X className="w-3 h-3" />
                      </button>
                    </span>
                  )}

                  {searchQuery && (
                    <span className="inline-flex items-center gap-1 px-3 py-1 bg-purple-100 text-purple-700 rounded-full text-sm">
                      Search: "{searchQuery}"
                      <button onClick={clearSearch} className="hover:text-red-500">
                        <X className="w-3 h-3" />
                      </button>
                    </span>
                  )}

                  {selectedBrand && (
                    <span className="inline-flex items-center gap-1 px-3 py-1 bg-pink-100 text-pink-700 rounded-full text-sm">
                      Brand: {selectedBrand}
                      <button onClick={() => setSelectedBrand('')} className="hover:text-red-500">
                        <X className="w-3 h-3" />
                      </button>
                    </span>
                  )}

                  {inStockOnly && (
                    <span className="inline-flex items-center gap-1 px-3 py-1 bg-green-100 text-green-700 rounded-full text-sm">
                      In Stock
                      <button onClick={() => setInStockOnly(false)} className="hover:text-red-500">
                        <X className="w-3 h-3" />
                      </button>
                    </span>
                  )}

                  {onSaleOnly && (
                    <span className="inline-flex items-center gap-1 px-3 py-1 bg-red-100 text-red-700 rounded-full text-sm">
                      On Sale
                      <button onClick={() => setOnSaleOnly(false)} className="hover:text-red-500">
                        <X className="w-3 h-3" />
                      </button>
                    </span>
                  )}

                  {minRating > 0 && (
                    <span className="inline-flex items-center gap-1 px-3 py-1 bg-yellow-100 text-yellow-700 rounded-full text-sm">
                      {minRating}+ Stars
                      <button onClick={() => setMinRating(0)} className="hover:text-red-500">
                        <X className="w-3 h-3" />
                      </button>
                    </span>
                  )}

                  {appliedPriceRange && (
                    <span className="inline-flex items-center gap-1 px-3 py-1 bg-orange-100 text-orange-700 rounded-full text-sm">
                      {formatPrice(appliedPriceRange[0])} - {formatPrice(appliedPriceRange[1])}
                      <button onClick={handleClearPriceFilter} className="hover:text-red-500">
                        <X className="w-3 h-3" />
                      </button>
                    </span>
                  )}

                  <button
                    onClick={clearAllFilters}
                    className="text-sm text-red-500 hover:underline ml-2"
                  >
                    Clear All
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
                    Filters
                    {hasActiveFilters && (
                      <span className="w-2 h-2 bg-[var(--color-primary)] rounded-full"></span>
                    )}
                  </button>

                  <p className="text-sm text-[var(--color-text-secondary)]">
                    Showing <strong>{products.length}</strong> of <strong>{totalProducts}</strong> products
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
                <h3 className="font-serif mb-2">No Products Found</h3>
                <p className="text-[var(--color-text-secondary)] mb-4">
                  Try adjusting your filters or search criteria
                </p>
                <button
                  onClick={clearAllFilters}
                  className="px-6 py-2 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-dark)] transition-colors"
                >
                  Clear All Filters
                </button>
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
                    Previous
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
                    Next
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
          <div className="absolute bottom-0 left-0 right-0 bg-white rounded-t-2xl max-h-[85vh] overflow-y-auto">
            <div className="p-4 border-b border-[var(--color-border)] flex items-center justify-between sticky top-0 bg-white z-10">
              <h3 className="font-serif">Filters</h3>
              <button onClick={() => setShowFilters(false)}>
                <X className="w-6 h-6" />
              </button>
            </div>

            <div className="p-4 space-y-6">
              {/* Categories */}
              <div>
                <h4 className="font-medium mb-3">Categories</h4>
                <div className="space-y-2">
                  <button
                    onClick={() => {
                      handleCategorySelect(null);
                      setShowFilters(false);
                    }}
                    className={`w-full text-left px-3 py-2 rounded-lg ${
                      !selectedCategoryId ? 'bg-[var(--color-primary)] text-white' : 'bg-[var(--color-surface)]'
                    }`}
                  >
                    All Products
                  </button>
                  {categories.map((category) => (
                    <button
                      key={category.id}
                      onClick={() => {
                        handleCategorySelect(category.id);
                        setShowFilters(false);
                      }}
                      className={`w-full text-left px-3 py-2 rounded-lg flex items-center gap-2 ${
                        selectedCategoryId === category.id
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

              {/* Quick Filters */}
              <div>
                <h4 className="font-medium mb-3">Quick Filters</h4>
                <div className="space-y-2">
                  <label className="flex items-center gap-2 cursor-pointer p-2 bg-[var(--color-surface)] rounded-lg">
                    <input
                      type="checkbox"
                      checked={inStockOnly}
                      onChange={(e) => setInStockOnly(e.target.checked)}
                      className="w-4 h-4 rounded"
                    />
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    <span>In Stock Only</span>
                  </label>
                  <label className="flex items-center gap-2 cursor-pointer p-2 bg-[var(--color-surface)] rounded-lg">
                    <input
                      type="checkbox"
                      checked={onSaleOnly}
                      onChange={(e) => setOnSaleOnly(e.target.checked)}
                      className="w-4 h-4 rounded"
                    />
                    <Percent className="w-4 h-4 text-red-500" />
                    <span>On Sale</span>
                  </label>
                </div>
              </div>

              {/* Rating Filter */}
              <div>
                <h4 className="font-medium mb-3">Minimum Rating</h4>
                <div className="flex flex-wrap gap-2">
                  {[0, 1, 2, 3, 4].map((rating) => (
                    <button
                      key={rating}
                      onClick={() => setMinRating(rating)}
                      className={`px-3 py-2 rounded-lg text-sm flex items-center gap-1 ${
                        minRating === rating
                          ? 'bg-[var(--color-primary)] text-white'
                          : 'bg-[var(--color-surface)]'
                      }`}
                    >
                      {rating > 0 ? (
                        <>
                          <Star className={`w-4 h-4 ${minRating === rating ? 'fill-current' : 'text-yellow-400 fill-current'}`} />
                          {rating}+
                        </>
                      ) : (
                        'All'
                      )}
                    </button>
                  ))}
                </div>
              </div>

              {/* Price Range */}
              <div>
                <h4 className="font-medium mb-3">Price Range</h4>
                <div className="space-y-2">
                  <div className="flex gap-2">
                    <input
                      type="number"
                      value={priceRange[0]}
                      onChange={(e) => setPriceRange([parseInt(e.target.value) || 0, priceRange[1]])}
                      placeholder="Min"
                      className="w-full px-3 py-2 text-sm border border-[var(--color-border)] rounded-lg"
                    />
                    <span className="text-[var(--color-text-muted)] self-center">-</span>
                    <input
                      type="number"
                      value={priceRange[1]}
                      onChange={(e) => setPriceRange([priceRange[0], parseInt(e.target.value) || 0])}
                      placeholder="Max"
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
                    Apply Price Filter
                  </button>
                </div>
              </div>

              {/* Brands */}
              {brands.length > 0 && (
                <div>
                  <h4 className="font-medium mb-3">Brands</h4>
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

              {/* Clear & Apply Buttons */}
              <div className="flex gap-3 pt-4 border-t border-[var(--color-border)]">
                {hasActiveFilters && (
                  <button
                    onClick={() => {
                      clearAllFilters();
                      setShowFilters(false);
                    }}
                    className="flex-1 py-3 text-red-500 border border-red-200 rounded-lg font-medium"
                  >
                    Clear All
                  </button>
                )}
                <button
                  onClick={() => setShowFilters(false)}
                  className="flex-1 py-3 bg-[var(--color-primary)] text-white rounded-lg font-medium"
                >
                  Apply Filters
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
