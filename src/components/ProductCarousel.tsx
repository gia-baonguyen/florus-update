import { ChevronLeft, ChevronRight } from 'lucide-react';
import { useRef, useState } from 'react';
import { ProductCard } from './ProductCard';
import type { Product } from '../types';

interface ProductCarouselProps {
  title: string;
  subtitle?: string;
  products: Product[];
  onProductClick: (productId: number) => void;
  recommendationLabel?: string;
}

export function ProductCarousel({
  title,
  subtitle,
  products,
  onProductClick,
  recommendationLabel,
}: ProductCarouselProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const [canScrollLeft, setCanScrollLeft] = useState(false);
  const [canScrollRight, setCanScrollRight] = useState(true);

  const checkScroll = () => {
    if (scrollRef.current) {
      const { scrollLeft, scrollWidth, clientWidth } = scrollRef.current;
      setCanScrollLeft(scrollLeft > 0);
      setCanScrollRight(scrollLeft < scrollWidth - clientWidth - 10);
    }
  };

  const scroll = (direction: 'left' | 'right') => {
    if (scrollRef.current) {
      const scrollAmount = 300;
      const newScrollLeft =
        direction === 'left'
          ? scrollRef.current.scrollLeft - scrollAmount
          : scrollRef.current.scrollLeft + scrollAmount;

      scrollRef.current.scrollTo({
        left: newScrollLeft,
        behavior: 'smooth',
      });

      setTimeout(checkScroll, 300);
    }
  };

  return (
    <div className="bg-white rounded-xl border border-[var(--color-border)] p-5 md:p-6 shadow-[var(--shadow-sm)]">
      {/* Header */}
      <div className="flex items-center justify-between mb-5">
        <div>
          <h2 className="text-[var(--color-text-primary)] font-serif">{title}</h2>
          {subtitle && (
            <p className="text-[var(--color-text-secondary)] mt-1 text-sm">{subtitle}</p>
          )}
        </div>

        {/* Navigation Buttons - Desktop */}
        <div className="hidden md:flex items-center gap-2">
          <button
            onClick={() => scroll('left')}
            disabled={!canScrollLeft}
            className="p-2.5 rounded-full border border-[var(--color-border)] hover:bg-[var(--color-primary-light)] hover:border-[var(--color-primary)] disabled:opacity-30 disabled:cursor-not-allowed transition-all"
            aria-label="Scroll left"
          >
            <ChevronLeft className="w-5 h-5" />
          </button>
          <button
            onClick={() => scroll('right')}
            disabled={!canScrollRight}
            className="p-2.5 rounded-full border border-[var(--color-border)] hover:bg-[var(--color-primary-light)] hover:border-[var(--color-primary)] disabled:opacity-30 disabled:cursor-not-allowed transition-all"
            aria-label="Scroll right"
          >
            <ChevronRight className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Products Scroll Container */}
      <div
        ref={scrollRef}
        onScroll={checkScroll}
        className="flex gap-4 overflow-x-auto scrollbar-hide snap-x snap-mandatory"
        style={{ scrollbarWidth: 'none', msOverflowStyle: 'none' }}
      >
        {products.map((product) => (
          <div
            key={product.id}
            className="flex-none w-[180px] md:w-[220px] snap-start"
          >
            <ProductCard
              product={product}
              onClick={() => onProductClick(product.id)}
              compact
            />
          </div>
        ))}
      </div>

      {/* Mobile scroll hint */}
      <div className="md:hidden mt-4 flex justify-center">
        <p className="text-xs text-[var(--color-text-muted)] italic">
          ← Vuốt để xem thêm →
        </p>
      </div>
    </div>
  );
}