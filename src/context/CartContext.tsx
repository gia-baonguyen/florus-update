import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';
import { Cart, CartItem, Product } from '../types';
import { cartApi, recommendationsApi } from '../api/products';
import { useAuth } from './AuthContext';

interface CartContextType {
  cart: Cart | null;
  loading: boolean;
  error: string | null;
  crossSellProducts: Product[];
  addItem: (productId: number, quantity?: number) => Promise<void>;
  updateItem: (productId: number, quantity: number) => Promise<void>;
  removeItem: (productId: number) => Promise<void>;
  clearCart: () => Promise<void>;
  refreshCart: () => Promise<void>;
}

const CartContext = createContext<CartContextType | undefined>(undefined);

export const CartProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [cart, setCart] = useState<Cart | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [crossSellProducts, setCrossSellProducts] = useState<Product[]>([]);
  const { isAuthenticated } = useAuth();

  const refreshCart = useCallback(async () => {
    if (!isAuthenticated) {
      setCart(null);
      setCrossSellProducts([]);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      const cartData = await cartApi.get();
      setCart(cartData);

      // Fetch cross-sell recommendations
      if (cartData.items && cartData.items.length > 0) {
        const productId = cartData.items[0].product_id;
        try {
          const crossSell = await recommendationsApi.getCrossSell(productId);
          // Filter out items already in cart
          const cartProductIds = cartData.items.map(item => item.product_id);
          const filtered = crossSell.filter(p => !cartProductIds.includes(p.id));
          setCrossSellProducts(filtered.slice(0, 3));
        } catch {
          setCrossSellProducts([]);
        }
      } else {
        setCrossSellProducts([]);
      }
    } catch (err) {
      console.error('Error fetching cart:', err);
      setError('Không thể tải giỏ hàng');
      setCart(null);
    } finally {
      setLoading(false);
    }
  }, [isAuthenticated]);

  useEffect(() => {
    refreshCart();
  }, [refreshCart]);

  const addItem = async (productId: number, quantity = 1) => {
    if (!isAuthenticated) {
      throw new Error('Vui lòng đăng nhập để thêm sản phẩm vào giỏ hàng');
    }

    try {
      setLoading(true);
      setError(null);
      const updatedCart = await cartApi.addItem(productId, quantity);
      setCart(updatedCart);
    } catch (err) {
      console.error('Error adding to cart:', err);
      setError('Không thể thêm sản phẩm vào giỏ hàng');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const updateItem = async (productId: number, quantity: number) => {
    if (!isAuthenticated) return;

    try {
      setLoading(true);
      setError(null);
      const updatedCart = await cartApi.updateItem(productId, quantity);
      setCart(updatedCart);
    } catch (err) {
      console.error('Error updating cart:', err);
      setError('Không thể cập nhật giỏ hàng');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const removeItem = async (productId: number) => {
    if (!isAuthenticated) return;

    try {
      setLoading(true);
      setError(null);
      await cartApi.removeItem(productId);
      await refreshCart();
    } catch (err) {
      console.error('Error removing from cart:', err);
      setError('Không thể xóa sản phẩm khỏi giỏ hàng');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const clearCart = async () => {
    if (!isAuthenticated) return;

    try {
      setLoading(true);
      setError(null);
      await cartApi.clear();
      setCart(null);
      setCrossSellProducts([]);
    } catch (err) {
      console.error('Error clearing cart:', err);
      setError('Không thể xóa giỏ hàng');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return (
    <CartContext.Provider
      value={{
        cart,
        loading,
        error,
        crossSellProducts,
        addItem,
        updateItem,
        removeItem,
        clearCart,
        refreshCart,
      }}
    >
      {children}
    </CartContext.Provider>
  );
};

export const useCart = () => {
  const context = useContext(CartContext);
  if (context === undefined) {
    throw new Error('useCart must be used within a CartProvider');
  }
  return context;
};

export default CartContext;
