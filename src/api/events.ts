import api from './client';

// Session ID management
const SESSION_KEY = 'florus_session_id';

function getSessionId(): string {
  let sessionId = sessionStorage.getItem(SESSION_KEY);
  if (!sessionId) {
    sessionId = crypto.randomUUID();
    sessionStorage.setItem(SESSION_KEY, sessionId);
  }
  return sessionId;
}

// Add session ID to all event requests
function getHeaders() {
  return {
    'X-Session-ID': getSessionId(),
  };
}

export interface EventData {
  event_type: string;
  product_id?: number;
  category_id?: number;
  order_id?: number;
  search_query?: string;
  quantity?: number;
  price?: number;
  rating?: number;
  metadata?: string;
}

export const eventsApi = {
  // Generic event logging
  logEvent: async (data: EventData) => {
    try {
      const response = await api.post('/events', data, { headers: getHeaders() });
      return response.data;
    } catch (error) {
      console.error('Failed to log event:', error);
      // Don't throw - event logging should not break the app
    }
  },

  // Product view event
  logProductView: async (productId: number, categoryId?: number) => {
    return eventsApi.logEvent({
      event_type: 'product_view',
      product_id: productId,
      category_id: categoryId,
    });
  },

  // Product click event
  logProductClick: async (productId: number) => {
    return eventsApi.logEvent({
      event_type: 'product_click',
      product_id: productId,
    });
  },

  // Add to cart event
  logAddToCart: async (productId: number, quantity: number, price: number) => {
    return eventsApi.logEvent({
      event_type: 'add_to_cart',
      product_id: productId,
      quantity,
      price,
    });
  },

  // Remove from cart event
  logRemoveFromCart: async (productId: number) => {
    return eventsApi.logEvent({
      event_type: 'remove_from_cart',
      product_id: productId,
    });
  },

  // Purchase event (called for each item in order)
  logPurchase: async (productId: number, orderId: number, quantity: number, price: number) => {
    return eventsApi.logEvent({
      event_type: 'purchase',
      product_id: productId,
      order_id: orderId,
      quantity,
      price,
    });
  },

  // Search event
  logSearch: async (query: string) => {
    return eventsApi.logEvent({
      event_type: 'search',
      search_query: query,
    });
  },

  // Category view event
  logCategoryView: async (categoryId: number) => {
    return eventsApi.logEvent({
      event_type: 'category_view',
      category_id: categoryId,
    });
  },

  // Add to wishlist event
  logAddToWishlist: async (productId: number) => {
    return eventsApi.logEvent({
      event_type: 'add_to_wishlist',
      product_id: productId,
    });
  },

  // Rating event
  logRating: async (productId: number, rating: number) => {
    return eventsApi.logEvent({
      event_type: 'rating',
      product_id: productId,
      rating,
    });
  },

  // Review event
  logReview: async (productId: number, rating: number) => {
    return eventsApi.logEvent({
      event_type: 'review',
      product_id: productId,
      rating,
    });
  },

  // Get popular products based on events
  getPopularProducts: async (limit: number = 10, days: number = 7) => {
    const response = await api.get('/events/popular-products', {
      params: { limit, days },
    });
    return response.data;
  },

  // Get user behavior profile (admin)
  getUserBehavior: async (userId: number) => {
    const response = await api.get(`/events/user/${userId}/behavior`);
    return response.data;
  },

  // Get product event stats (admin)
  getProductStats: async (productId: number) => {
    const response = await api.get(`/events/product/${productId}/stats`);
    return response.data;
  },
};

export default eventsApi;
