import { useState, useEffect, useRef } from 'react';
import { Search, ShoppingCart, User, Sparkles, LogOut, X, Loader2, History, Heart } from 'lucide-react';
import { Link, useLocation } from 'react-router-dom';
import { User as UserType, Product } from '../types';
import { productsApi } from '../api/products';
import { ProductImage } from './ProductImage';

interface HeaderProps {
  cartCount: number;
  onCartClick: () => void;
  onLogoClick: () => void;
  user?: UserType | null;
  onLogout?: () => void;
  onProductClick?: (productId: string) => void;
}

export function Header({ cartCount, onCartClick, onLogoClick, user, onLogout, onProductClick }: HeaderProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<Product[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [showResults, setShowResults] = useState(false);
  
  const location = useLocation();
  const isOrderHistoryPage = location.pathname === '/orders';
  const isWishlistPage = location.pathname === '/wishlist';
  
  const searchRef = useRef<HTMLDivElement>(null);
  const mobileSearchRef = useRef<HTMLDivElement>(null);

  // Debounced search
  useEffect(() => {
    if (searchQuery.trim().length < 2) {
      setSearchResults([]);
      setShowResults(false);
      return;
    }

    const timer = setTimeout(async () => {
      setIsSearching(true);
      try {
        const { products } = await productsApi.getAll(1, 8, undefined, searchQuery);
        setSearchResults(products);
        setShowResults(true);
      } catch (err) {
        console.error('Search error:', err);
        setSearchResults([]);
      } finally {
        setIsSearching(false);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery]);

  // Close on click outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        searchRef.current && !searchRef.current.contains(event.target as Node) &&
        mobileSearchRef.current && !mobileSearchRef.current.contains(event.target as Node)
      ) {
        setShowResults(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleProductSelect = (productId: number) => {
    setSearchQuery('');
    setShowResults(false);
    onProductClick?.(productId.toString());
  };

  const clearSearch = () => {
    setSearchQuery('');
    setSearchResults([]);
    setShowResults(false);
  };

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price / 1000);
  };

  const SearchResultsDropdown = () => (
    <div className="absolute top-full left-0 right-0 mt-1 bg-white border border-[var(--color-border)] rounded-xl shadow-[var(--shadow-lg)] max-h-96 overflow-y-auto z-50">
      {searchResults.length > 0 ? (
        <div className="p-2">
          {searchResults.map((product) => (
            <button
              key={product.id}
              onClick={() => handleProductSelect(product.id)}
              className="w-full flex items-center gap-3 p-2 hover:bg-[var(--color-surface-warm)] rounded-lg transition-colors text-left"
            >
              <div className="w-12 h-12 rounded-lg overflow-hidden bg-[var(--color-surface-warm)] flex-shrink-0 border border-[var(--color-border)]">
                <ProductImage imageUrl={product.image_url} alt={product.name} size="small" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-xs text-[var(--color-text-muted)] uppercase">{product.brand}</p>
                <p className="text-sm font-medium truncate">{product.name}</p>
                <p className="text-sm text-[var(--color-primary)]">{formatPrice(product.price)}</p>
              </div>
            </button>
          ))}
        </div>
      ) : (
        <div className="p-6 text-center">
          <p className="text-[var(--color-text-muted)]">No products found</p>
        </div>
      )}
    </div>
  );

  return (
    <header className="bg-white border-b border-gray-100 sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <button onClick={onLogoClick} className="flex items-center gap-2 hover:opacity-80 transition-opacity">
            <Sparkles className="w-6 h-6 text-[var(--color-primary)]" />
            <h1 className="text-gray-900 font-serif">Florus Beauty</h1>
          </button>

          {/* Search Bar - Hidden on mobile */}
          <div ref={searchRef} className="hidden md:flex flex-1 max-w-md mx-8 relative">
            <div className="relative w-full">
              {isSearching ? (
                <Loader2 className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--color-primary)] animate-spin" />
              ) : (
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
              )}
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onFocus={() => searchResults.length > 0 && setShowResults(true)}
                placeholder="Search products..."
                className="w-full pl-10 pr-10 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)] focus:border-transparent"
              />
              {searchQuery && (
                <button
                  onClick={clearSearch}
                  className="absolute right-3 top-1/2 -translate-y-1/2 p-1 hover:bg-gray-100 rounded-full transition-colors"
                >
                  <X className="w-4 h-4 text-gray-400" />
                </button>
              )}
            </div>
            {showResults && <SearchResultsDropdown />}
          </div>

          {/* Actions */}
          <div className="flex items-center gap-2">
            {user ? (
              <div className="flex items-center gap-2">
                {/* Wishlist Button */}
                <Link
                  to="/wishlist"
                  className={`p-2 rounded-lg transition-colors group ${
                    isWishlistPage
                      ? 'bg-[var(--color-primary-light)] text-[var(--color-primary)]'
                      : 'hover:bg-[var(--color-primary-light)] text-gray-500'
                  }`}
                  title="Wishlist"
                >
                  <Heart
                    className={`w-5 h-5 ${
                      isWishlistPage ? 'text-[var(--color-primary)]' : 'group-hover:text-[var(--color-primary)]'
                    }`}
                  />
                </Link>

                {/* Order History Button */}
                <Link
                  to="/orders"
                  className={`p-2 rounded-lg transition-colors group ${
                    isOrderHistoryPage 
                      ? 'bg-[var(--color-primary-light)] text-[var(--color-primary)]' 
                      : 'hover:bg-[var(--color-primary-light)] text-gray-500'
                  }`}
                  title="Order History"
                >
                  <History className={`w-5 h-5 ${
                    isOrderHistoryPage ? 'text-[var(--color-primary)]' : 'group-hover:text-[var(--color-primary)]'
                  }`} />
                </Link>

                {/* User Profile */}
                <div className="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-[var(--color-primary-light)] rounded-full">
                  <div className="w-6 h-6 bg-[var(--color-primary)] rounded-full flex items-center justify-center text-white text-xs">
                    {user.name.charAt(0).toUpperCase()}
                  </div>
                  <span className="text-sm text-[var(--color-primary-dark)] font-medium">{user.name}</span>
                </div>
                
                {/* Logout Button */}
                <button
                  onClick={onLogout}
                  className="p-2 hover:bg-red-50 rounded-lg transition-colors group"
                  title="Sign Out"
                >
                  <LogOut className="w-5 h-5 text-gray-500 group-hover:text-red-500" />
                </button>
              </div>
            ) : (
              <Link
                to="/login"
                className="flex items-center gap-2 px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-dark)] transition-colors"
              >
                <User className="w-4 h-4" />
                <span className="hidden sm:inline text-sm font-medium">Sign In</span>
              </Link>
            )}

            {/* Shopping Cart Button */}
            <button
              onClick={onCartClick}
              className="relative p-2 hover:bg-gray-50 rounded-lg transition-colors"
            >
              <ShoppingCart className="w-5 h-5 text-gray-700" />
              {cartCount > 0 && (
                <span className="absolute -top-1 -right-1 bg-[var(--color-primary)] text-white text-xs w-5 h-5 flex items-center justify-center rounded-full">
                  {cartCount}
                </span>
              )}
            </button>
          </div>
        </div>
      </div>

      {/* Mobile Search */}
      <div ref={mobileSearchRef} className="md:hidden px-4 pb-3 relative">
        <div className="relative">
          {isSearching ? (
            <Loader2 className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--color-primary)] animate-spin" />
          ) : (
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
          )}
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onFocus={() => searchResults.length > 0 && setShowResults(true)}
            placeholder="Search products..."
            className="w-full pl-10 pr-10 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)] focus:border-transparent"
          />
          {searchQuery && (
            <button
              onClick={clearSearch}
              className="absolute right-3 top-1/2 -translate-y-1/2 p-1 hover:bg-gray-100 rounded-full transition-colors"
            >
              <X className="w-4 h-4 text-gray-400" />
            </button>
          )}
        </div>
        {showResults && <SearchResultsDropdown />}
      </div>
    </header>
  );
}