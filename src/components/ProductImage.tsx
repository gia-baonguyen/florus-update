import { ImageWithFallback } from './figma/ImageWithFallback';

interface ProductImageProps {
  imageUrl: string;
  alt: string;
  size?: 'small' | 'medium' | 'large';
}

export function ProductImage({ imageUrl, alt, size = 'medium' }: ProductImageProps) {
  const sizeClasses = {
    small: 'w-20 h-20',
    medium: 'w-full h-48',
    large: 'w-full h-96'
  };

  return (
    <div className={`${sizeClasses[size]} overflow-hidden bg-gray-50 rounded-lg`}>
      <ImageWithFallback
        src={imageUrl}
        alt={alt}
        className="w-full h-full object-cover hover:scale-105 transition-transform duration-300"
      />
    </div>
  );
}