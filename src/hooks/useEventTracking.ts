import { useCallback, useEffect, useRef } from 'react';
import { eventsApi } from '../api/events';

/**
 * Hook for tracking user events in components
 */
export function useEventTracking() {
  // Track product view with debounce to avoid duplicate events
  const viewedProducts = useRef<Set<number>>(new Set());

  const trackProductView = useCallback((productId: number, categoryId?: number) => {
    // Only track once per session per product
    if (viewedProducts.current.has(productId)) {
      return;
    }
    viewedProducts.current.add(productId);
    eventsApi.logProductView(productId, categoryId);
  }, []);

  const trackProductClick = useCallback((productId: number) => {
    eventsApi.logProductClick(productId);
  }, []);

  const trackAddToCart = useCallback((productId: number, quantity: number, price: number) => {
    eventsApi.logAddToCart(productId, quantity, price);
  }, []);

  const trackRemoveFromCart = useCallback((productId: number) => {
    eventsApi.logRemoveFromCart(productId);
  }, []);

  const trackPurchase = useCallback((productId: number, orderId: number, quantity: number, price: number) => {
    eventsApi.logPurchase(productId, orderId, quantity, price);
  }, []);

  const trackSearch = useCallback((query: string) => {
    if (query.trim().length >= 2) {
      eventsApi.logSearch(query.trim());
    }
  }, []);

  const trackCategoryView = useCallback((categoryId: number) => {
    eventsApi.logCategoryView(categoryId);
  }, []);

  const trackAddToWishlist = useCallback((productId: number) => {
    eventsApi.logAddToWishlist(productId);
  }, []);

  const trackRating = useCallback((productId: number, rating: number) => {
    eventsApi.logRating(productId, rating);
  }, []);

  const trackReview = useCallback((productId: number, rating: number) => {
    eventsApi.logReview(productId, rating);
  }, []);

  return {
    trackProductView,
    trackProductClick,
    trackAddToCart,
    trackRemoveFromCart,
    trackPurchase,
    trackSearch,
    trackCategoryView,
    trackAddToWishlist,
    trackRating,
    trackReview,
  };
}

/**
 * Hook for automatically tracking product page views
 */
export function useProductViewTracking(productId: number | undefined, categoryId?: number) {
  const { trackProductView } = useEventTracking();

  useEffect(() => {
    if (productId) {
      trackProductView(productId, categoryId);
    }
  }, [productId, categoryId, trackProductView]);
}

/**
 * Hook for automatically tracking category page views
 */
export function useCategoryViewTracking(categoryId: number | undefined) {
  const { trackCategoryView } = useEventTracking();

  useEffect(() => {
    if (categoryId) {
      trackCategoryView(categoryId);
    }
  }, [categoryId, trackCategoryView]);
}

/**
 * Hook for tracking search with debounce
 */
export function useSearchTracking(debounceMs: number = 1000) {
  const { trackSearch } = useEventTracking();
  const timeoutRef = useRef<NodeJS.Timeout>();
  const lastQueryRef = useRef<string>('');

  const trackSearchDebounced = useCallback((query: string) => {
    // Clear previous timeout
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    // Don't track same query twice
    if (query === lastQueryRef.current) {
      return;
    }

    timeoutRef.current = setTimeout(() => {
      if (query.trim().length >= 2) {
        lastQueryRef.current = query;
        trackSearch(query);
      }
    }, debounceMs);
  }, [trackSearch, debounceMs]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  return trackSearchDebounced;
}

export default useEventTracking;
