export interface Product {
  id: string;
  name: string;
  price: number;
  originalPrice?: number;
  brand: string;
  rating: number;
  reviewCount: number;
  category: string;
  imageUrl: string;
  aiRecommendType: string;
  aiScore: number;
  description: string;
  ingredients?: string[];
  tags: string[];
}

export const products: Product[] = [
  {
    id: '1',
    name: 'Vitamin C Brightening Serum',
    brand: 'Radiance Labs',
    price: 580000,
    originalPrice: 720000,
    rating: 4.8,
    reviewCount: 342,
    category: 'Serum',
    imageUrl: 'https://images.unsplash.com/photo-1751131964776-57e3cbca0a14?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxsdXh1cnklMjBza2luY2FyZSUyMHNlcnVtJTIwYm90dGxlfGVufDF8fHx8MTc2NzM0NDI2N3ww&ixlib=rb-4.1.0&q=80&w=1080',
    aiRecommendType: 'cold-start',
    aiScore: 95,
    description: 'Serum vitamin C 20% giúp làm sáng và đều màu da',
    ingredients: ['Vitamin C 20%', 'Ferulic Acid', 'Vitamin E'],
    tags: ['Làm sáng', 'Chống oxy hóa', 'Mờ thâm'],
  },
  {
    id: '2',
    name: 'Hyaluronic Acid Moisturizer',
    brand: 'Aqua Glow',
    price: 420000,
    rating: 4.6,
    reviewCount: 256,
    category: 'Moisturizer',
    imageUrl: 'https://images.unsplash.com/photo-1763503834047-ac85c4105c0b?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxuYXR1cmFsJTIwZmFjZSUyMGNyZWFtJTIwamFyfGVufDF8fHx8MTc2NzQxMTA4Mnww&ixlib=rb-4.1.0&q=80&w=1080',
    aiRecommendType: 'warm-start',
    aiScore: 92,
    description: 'Kem dưỡng ẩm với Hyaluronic Acid cấp nước 24h',
    ingredients: ['Hyaluronic Acid', 'Ceramide', 'Niacinamide'],
    tags: ['Cấp ẩm', 'Phục hồi', 'Không nhờn'],
  },
  {
    id: '3',
    name: 'Matte Lipstick - Coral Dream',
    brand: 'Velvet Kiss',
    price: 290000,
    rating: 4.7,
    reviewCount: 189,
    category: 'Lipstick',
    imageUrl: 'https://images.unsplash.com/photo-1762522919970-18818d127017?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxlbGVnYW50JTIwbGlwc3RpY2slMjBtYWtldXB8ZW58MXx8fHwxNzY3NDExMDgyfDA&ixlib=rb-4.1.0&q=80&w=1080',
    aiRecommendType: 'collaborative',
    aiScore: 88,
    description: 'Son lì mịn môi, giữ màu 12h',
    tags: ['Lâu trôi', 'Mịn môi', 'Không khô'],
  },
  {
    id: '4',
    name: 'Retinol Eye Cream',
    brand: 'Youth Restore',
    price: 650000,
    originalPrice: 850000,
    rating: 4.9,
    reviewCount: 423,
    category: 'Eye Cream',
    imageUrl: 'https://images.unsplash.com/photo-1648712789205-4a05ebb8d026?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxwcmVtaXVtJTIwZXllJTIwY3JlYW18ZW58MXx8fHwxNzY3NDExMDgzfDA&ixlib=rb-4.1.0&q=80&w=1080',
    aiRecommendType: 'content-based',
    aiScore: 96,
    description: 'Kem mắt Retinol giảm thâm quầng và nếp nhăn',
    ingredients: ['Retinol 0.5%', 'Peptides', 'Caffeine'],
    tags: ['Chống lão hóa', 'Giảm thâm', 'Săn chắc'],
  },
  {
    id: '5',
    name: 'Flawless Foundation SPF 30',
    brand: 'Perfect Canvas',
    price: 520000,
    rating: 4.5,
    reviewCount: 178,
    category: 'Foundation',
    imageUrl: 'https://images.unsplash.com/photo-1670065921509-d0e0d053355b?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxsdXh1cnklMjBmb3VuZGF0aW9uJTIwYm90dGxlfGVufDF8fHx8MTc2NzQxMTA4M3ww&ixlib=rb-4.1.0&q=80&w=1080',
    aiRecommendType: 'cross-sell',
    aiScore: 87,
    description: 'Kem nền che phủ hoàn hảo, tự nhiên với SPF 30',
    tags: ['Che phủ cao', 'SPF 30', 'Tự nhiên'],
  },
  {
    id: '6',
    name: 'Gentle Foam Cleanser',
    brand: 'Pure Clean',
    price: 280000,
    rating: 4.6,
    reviewCount: 312,
    category: 'Cleanser',
    imageUrl: 'https://images.unsplash.com/photo-1763622499218-37fdfc7a590a?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxuYXR1cmFsJTIwY2xlYW5zZXIlMjBib3R0bGV8ZW58MXx8fHwxNzY3NDExMDg0fDA&ixlib=rb-4.1.0&q=80&w=1080',
    aiRecommendType: 'cold-start',
    aiScore: 90,
    description: 'Sữa rửa mặt dịu nhẹ, làm sạch sâu không khô da',
    ingredients: ['Amino Acids', 'Glycerin', 'Aloe Vera'],
    tags: ['Dịu nhẹ', 'Làm sạch sâu', 'pH cân bằng'],
  },
  {
    id: '7',
    name: 'Volume Mascara - Midnight Black',
    brand: 'Lash Perfect',
    price: 320000,
    rating: 4.7,
    reviewCount: 267,
    category: 'Mascara',
    imageUrl: 'https://images.unsplash.com/photo-1586361984196-9cb94b4fff55?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxwcmVtaXVtJTIwbWFzY2FyYXxlbnwxfHx8fDE3Njc0MTEwODR8MA&ixlib=rb-4.1.0&q=80&w=1080',
    aiRecommendType: 'collaborative',
    aiScore: 85,
    description: 'Mascara dày mi, cong vút tự nhiên',
    tags: ['Dày mi', 'Cong vút', 'Không lem'],
  },
  {
    id: '8',
    name: 'Baked Blush - Rose Petal',
    brand: 'Glow Studio',
    price: 350000,
    rating: 4.8,
    reviewCount: 201,
    category: 'Blush',
    imageUrl: 'https://images.unsplash.com/photo-1674672524653-ebfea176496b?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxsdXh1cnklMjBibHVzaCUyMGNvbXBhY3R8ZW58MXx8fHwxNzY3NDExMDg0fDA&ixlib=rb-4.1.0&q=80&w=1080',
    aiRecommendType: 'content-based',
    aiScore: 89,
    description: 'Phấn má hồng dạng nướng, tự nhiên rạng rỡ',
    tags: ['Tự nhiên', 'Lâu trôi', 'Dễ tán'],
  },
  {
    id: '9',
    name: 'Eau de Parfum - Garden Rose',
    brand: 'Essence Luxe',
    price: 980000,
    originalPrice: 1200000,
    rating: 4.9,
    reviewCount: 156,
    category: 'Perfume',
    imageUrl: 'https://images.unsplash.com/photo-1747052881000-a640a4981dd0?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxlbGVnYW50JTIwcGVyZnVtZSUyMGJvdHRsZXxlbnwxfHx8fDE3Njc0MTEwODV8MA&ixlib=rb-4.1.0&q=80&w=1080',
    aiRecommendType: 'warm-start',
    aiScore: 93,
    description: 'Nước hoa hương hoa hồng thanh lịch, lưu hương 8h',
    tags: ['Sang trọng', 'Lưu hương lâu', 'Quyến rũ'],
  },
  {
    id: '10',
    name: 'Mineral Sunscreen SPF 50+',
    brand: 'Sun Shield',
    price: 380000,
    rating: 4.7,
    reviewCount: 289,
    category: 'Sunscreen',
    imageUrl: 'https://images.unsplash.com/photo-1636655104827-b27d8b6c05ac?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxuYXR1cmFsJTIwc3Vuc2NyZWVufGVufDF8fHx8MTc2NzQxMTA4NXww&ixlib=rb-4.1.0&q=80&w=1080',
    aiRecommendType: 'cross-sell',
    aiScore: 91,
    description: 'Kem chống nắng khoáng SPF 50+ PA++++, không trắng bệt',
    ingredients: ['Zinc Oxide', 'Titanium Dioxide', 'Niacinamide'],
    tags: ['SPF 50+', 'Không trắng bệt', 'Không gây mụn'],
  },
  {
    id: '11',
    name: 'Clay Detox Mask',
    brand: 'Skin Therapy',
    price: 450000,
    rating: 4.6,
    reviewCount: 234,
    category: 'Mask',
    imageUrl: 'https://images.unsplash.com/photo-1597294718483-9cf386211d3d?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxsdXh1cnklMjBmYWNlJTIwbWFza3xlbnwxfHx8fDE3Njc0MTEwODV8MA&ixlib=rb-4.1.0&q=80&w=1080',
    aiRecommendType: 'content-based',
    aiScore: 86,
    description: 'Mặt nạ đất sét thải độc, thu nhỏ lỗ chân lông',
    ingredients: ['Kaolin Clay', 'Charcoal', 'Tea Tree Oil'],
    tags: ['Thải độc', 'Thu nhỏ lỗ chân lông', 'Kiểm soát dầu'],
  },
  {
    id: '12',
    name: 'Gel Nail Polish - Nude Rose',
    brand: 'Nail Couture',
    price: 180000,
    rating: 4.5,
    reviewCount: 145,
    category: 'Nail Polish',
    imageUrl: 'https://images.unsplash.com/photo-1600852306771-c963331af110?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxwcmVtaXVtJTIwbmFpbCUyMHBvbGlzaHxlbnwxfHx8fDE3Njc0MTEwODZ8MA&ixlib=rb-4.1.0&q=80&w=1080',
    aiRecommendType: 'collaborative',
    aiScore: 84,
    description: 'Sơn gel móng tay, bóng mượt lâu trôi',
    tags: ['Bóng mượt', 'Lâu trôi', 'Dễ thoa'],
  },
  {
    id: '13',
    name: 'Eyeshadow Palette - Natural Nude',
    brand: 'Color Art',
    price: 620000,
    originalPrice: 780000,
    rating: 4.8,
    reviewCount: 298,
    category: 'Eyeshadow',
    imageUrl: 'https://images.unsplash.com/photo-1633434679701-9767cd2ba15a?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxlbGVnYW50JTIwZXllc2hhZG93JTIwcGFsZXR0ZXxlbnwxfHx8fDE3Njc0MTEwODZ8MA&ixlib=rb-4.1.0&q=80&w=1080',
    aiRecommendType: 'warm-start',
    aiScore: 94,
    description: 'Bảng phấn mắt 12 màu nude, dễ tán',
    tags: ['Đa năng', 'Dễ tán', 'Bền màu'],
  },
];

// Mock recommendation data based on Amazon All_Beauty dataset patterns
export const coldStartRecommendations = [
  "1", "2", "7", "8", "9", "3" // Top rated & trending beauty products
];

export const warmStartRecommendations = [
  "2", "5", "10", "13" // Personalized skincare recommendations
];

// Similar products - Content-based filtering (same category, brand, or purpose)
export const similarProducts = {
  "1": ["12", "9", "11"], // Lipstick -> Lip liner, Blush, Setting spray
  "2": ["5", "10", "13"], // Face cream -> Serum, Night cream, Cleanser
  "3": ["9", "11"], // Eyeshadow -> Blush, Setting spray
  "4": [], // Perfume is unique
  "5": ["2", "10", "13"], // Serum -> Cream, Night cream, Cleanser
  "6": [], // Nail polish
  "7": ["11", "3"], // Mascara -> Setting spray, Eyeshadow
  "8": ["2", "5"], // Face mask -> Cream, Serum
  "9": ["1", "3"], // Blush -> Lipstick, Eyeshadow
  "10": ["2", "5"], // Night cream -> Day cream, Serum
  "11": ["7", "3"], // Setting spray -> Mascara, Eyeshadow
  "12": ["1", "9"], // Lip liner -> Lipstick, Blush
  "13": ["2", "5", "8"], // Cleanser -> Cream, Serum, Mask
};

// Co-viewed products - Collaborative filtering (Graph-based relationships)
export const coViewedProducts = {
  "1": ["12", "11", "9"], // Lipstick often viewed with liner, setting spray, blush
  "2": ["5", "13", "8"], // Face cream with serum, cleanser, mask
  "3": ["7", "11", "9"], // Eyeshadow with mascara, setting spray, blush
  "4": [], // Perfume
  "5": ["2", "8", "13"], // Serum with cream, mask, cleanser
  "6": [], // Nail polish
  "7": ["3", "11"], // Mascara with eyeshadow, setting spray
  "8": ["2", "5", "13"], // Mask with cream, serum, cleanser
  "9": ["1", "3", "7"], // Blush with lipstick, eyeshadow, mascara
  "10": ["2", "5"], // Night cream with day cream, serum
  "11": ["1", "3", "7"], // Setting spray with lipstick, eyeshadow, mascara
  "12": ["1", "11"], // Lip liner with lipstick, setting spray
  "13": ["2", "5", "8"], // Cleanser with cream, serum, mask
};

// Cross-sell products - Frequently bought together
export const crossSellProducts = {
  "1": ["12", "11"], // Lipstick cross-sell: lip liner, setting spray
  "2": ["5", "13"], // Face cream cross-sell: serum, cleanser
  "3": ["7", "11"], // Eyeshadow cross-sell: mascara, setting spray
  "5": ["13", "8"], // Serum cross-sell: cleanser, mask
  "7": ["11"], // Mascara cross-sell: setting spray
  "8": ["13"], // Mask cross-sell: cleanser
  "9": ["1", "11"], // Blush cross-sell: lipstick, setting spray
  "10": ["5"], // Night cream cross-sell: serum
};