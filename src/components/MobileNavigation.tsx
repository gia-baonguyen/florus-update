import { Home, ShoppingCart, User, Grid, Sparkles } from 'lucide-react';

interface MobileNavigationProps {
  currentPage: 'home' | 'cart' | 'categories' | 'account';
  onNavigate: (page: 'home' | 'cart' | 'categories' | 'account') => void;
  cartCount: number;
}

export function MobileNavigation({ currentPage, onNavigate, cartCount }: MobileNavigationProps) {
  const navItems = [
    { id: 'home' as const, icon: Home, label: 'Home' },
    { id: 'categories' as const, icon: Grid, label: 'Categories' },
    { id: 'cart' as const, icon: ShoppingCart, label: 'Cart', badge: cartCount },
    { id: 'account' as const, icon: User, label: 'Account' },
  ];

  return (
    <nav className="md:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-[var(--color-border)] z-50 safe-area-bottom shadow-[var(--shadow-lg)]">
      <div className="grid grid-cols-4">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = currentPage === item.id;

          return (
            <button
              key={item.id}
              onClick={() => onNavigate(item.id)}
              className={`relative flex flex-col items-center justify-center py-2.5 transition-all ${
                isActive
                  ? 'text-[var(--color-primary)]'
                  : 'text-[var(--color-text-muted)] active:bg-[var(--color-surface)]'
              }`}
            >
              <div className="relative">
                <Icon className={`w-6 h-6 ${isActive ? 'scale-110' : ''} transition-transform`} />
                {item.badge !== undefined && item.badge > 0 && (
                  <span className="absolute -top-2 -right-2 bg-[var(--color-primary)] text-white rounded-full w-5 h-5 flex items-center justify-center text-[10px] shadow-[var(--shadow-md)]">
                    {item.badge > 9 ? '9+' : item.badge}
                  </span>
                )}
              </div>
              <span className={`text-[10px] mt-1 ${isActive ? '' : ''}`}>{item.label}</span>
              {isActive && (
                <div className="absolute top-0 left-1/2 -translate-x-1/2 w-12 h-1 bg-[var(--color-primary)] rounded-full" />
              )}
            </button>
          );
        })}
      </div>
    </nav>
  );
}