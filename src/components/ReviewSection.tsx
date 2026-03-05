import React, { useState, useEffect, useRef, useCallback } from 'react';
import { Loader2, MessageSquare, Trash2, Edit2, X, ImagePlus, XCircle, Camera, RefreshCw } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { reviewsApi, Review, ReviewListResponse } from '../api/reviews';
import { uploadApi } from '../api/upload';
import { StarRating } from './StarRating';

interface ReviewSectionProps {
  productId: number;
}

export function ReviewSection({ productId }: ReviewSectionProps) {
  const { isAuthenticated, user } = useAuth();
  const [reviewData, setReviewData] = useState<ReviewListResponse | null>(null);
  const [userReview, setUserReview] = useState<Review | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [editMode, setEditMode] = useState(false);

  const [formData, setFormData] = useState({
    rating: 5,
    comment: '',
  });
  const [selectedImages, setSelectedImages] = useState<string[]>([]); // Preview URLs (base64 or MinIO URLs)
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]); // File objects for upload
  const [uploadProgress, setUploadProgress] = useState<string>('');
  const fileInputRef = useRef<HTMLInputElement>(null);
  const cameraInputRef = useRef<HTMLInputElement>(null); // For mobile camera capture
  const MAX_IMAGES = 5;

  // Camera states
  const [showCamera, setShowCamera] = useState(false);
  const [cameraStream, setCameraStream] = useState<MediaStream | null>(null);
  const [facingMode, setFacingMode] = useState<'user' | 'environment'>('environment');
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);

  // Check if device is mobile or getUserMedia is unavailable (HTTP on mobile)
  const isMobileOrInsecure = () => {
    const isMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
    const isSecure = window.location.protocol === 'https:' || window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
    return isMobile || !isSecure || !navigator.mediaDevices?.getUserMedia;
  };

  useEffect(() => {
    loadReviews();
  }, [productId]);

  useEffect(() => {
    if (isAuthenticated) {
      loadUserReview();
    }
  }, [productId, isAuthenticated]);

  const loadReviews = async () => {
    try {
      setLoading(true);
      const data = await reviewsApi.getProductReviews(productId);
      setReviewData(data);
    } catch (err) {
      console.error('Failed to load reviews:', err);
    } finally {
      setLoading(false);
    }
  };

  const loadUserReview = async () => {
    try {
      const review = await reviewsApi.getUserReview(productId);
      setUserReview(review);
      if (review) {
        setFormData({
          rating: review.rating,
          comment: review.comment,
        });
        // Load existing images
        if (review.images && review.images.length > 0) {
          setSelectedImages(review.images.map(img => img.url));
        }
      }
    } catch {
      setUserReview(null);
    }
  };

  const handleImageSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (!files) return;

    const remaining = MAX_IMAGES - selectedImages.length;
    const filesToProcess = Array.from(files).slice(0, remaining);

    filesToProcess.forEach(file => {
      if (!file.type.startsWith('image/')) return;

      // Store the file object for upload
      setSelectedFiles(prev => {
        if (prev.length >= MAX_IMAGES) return prev;
        return [...prev, file];
      });

      // Create preview URL
      const reader = new FileReader();
      reader.onload = (event) => {
        const result = event.target?.result as string;
        setSelectedImages(prev => {
          if (prev.length >= MAX_IMAGES) return prev;
          return [...prev, result];
        });
      };
      reader.readAsDataURL(file);
    });

    // Reset input
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  // Handle camera capture from native input (for mobile)
  const handleCameraCapture = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (!files || files.length === 0) return;

    const file = files[0];
    if (!file.type.startsWith('image/')) return;

    if (selectedImages.length >= MAX_IMAGES) {
      setError(`Maximum ${MAX_IMAGES} images allowed`);
      return;
    }

    // Store the file object for upload
    setSelectedFiles(prev => [...prev, file]);

    // Create preview URL
    const reader = new FileReader();
    reader.onload = (event) => {
      const result = event.target?.result as string;
      setSelectedImages(prev => [...prev, result]);
    };
    reader.readAsDataURL(file);

    // Reset input
    if (cameraInputRef.current) {
      cameraInputRef.current.value = '';
    }
  };

  // Handle Take Photo button click
  const handleTakePhoto = () => {
    if (selectedImages.length >= MAX_IMAGES) {
      setError(`Maximum ${MAX_IMAGES} images allowed`);
      return;
    }

    // On mobile or insecure context, use native camera input
    if (isMobileOrInsecure()) {
      cameraInputRef.current?.click();
    } else {
      // On desktop with secure context, use getUserMedia
      startCamera();
    }
  };

  const removeImage = (index: number) => {
    setSelectedImages(prev => prev.filter((_, i) => i !== index));
    setSelectedFiles(prev => prev.filter((_, i) => i !== index));
  };

  // Camera functions
  const startCamera = useCallback(async () => {
    // First show the modal so video element is rendered
    setShowCamera(true);

    // Wait for DOM to update
    await new Promise(resolve => setTimeout(resolve, 100));

    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode },
        audio: false,
      });

      setCameraStream(stream);

      // Now connect stream to video element directly
      if (videoRef.current) {
        videoRef.current.srcObject = stream;
        try {
          await videoRef.current.play();
        } catch (playErr) {
          console.error('Video play failed:', playErr);
        }
      }
    } catch (err) {
      console.error('Failed to access camera:', err);
      setError('Could not access camera. Please check permissions.');
      setShowCamera(false);
    }
  }, [facingMode]);

  const stopCamera = useCallback(() => {
    if (cameraStream) {
      cameraStream.getTracks().forEach(track => track.stop());
      setCameraStream(null);
    }
    setShowCamera(false);
  }, [cameraStream]);

  const switchCamera = useCallback(async () => {
    stopCamera();
    setFacingMode(prev => prev === 'user' ? 'environment' : 'user');
  }, [stopCamera]);

  // Re-start camera when facing mode changes
  useEffect(() => {
    if (showCamera && !cameraStream) {
      startCamera();
    }
  }, [facingMode, showCamera, cameraStream, startCamera]);

  // Set video srcObject when camera modal is shown and stream is ready
  useEffect(() => {
    if (showCamera && cameraStream) {
      // Use requestAnimationFrame to ensure DOM is ready
      const setVideoSource = () => {
        if (videoRef.current) {
          videoRef.current.srcObject = cameraStream;
          videoRef.current.play().catch(err => {
            console.error('Failed to play video:', err);
          });
        } else {
          // Retry if video element not ready yet
          requestAnimationFrame(setVideoSource);
        }
      };
      requestAnimationFrame(setVideoSource);
    }
  }, [showCamera, cameraStream]);

  // Cleanup camera on unmount
  useEffect(() => {
    return () => {
      if (cameraStream) {
        cameraStream.getTracks().forEach(track => track.stop());
      }
    };
  }, [cameraStream]);

  const capturePhoto = useCallback(() => {
    if (!videoRef.current || !canvasRef.current) return;
    if (selectedImages.length >= MAX_IMAGES) {
      setError(`Maximum ${MAX_IMAGES} images allowed`);
      return;
    }

    const video = videoRef.current;
    const canvas = canvasRef.current;
    const context = canvas.getContext('2d');

    if (!context) return;

    // Set canvas size to video dimensions
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;

    // Draw the video frame to the canvas
    context.drawImage(video, 0, 0, canvas.width, canvas.height);

    // Convert to blob and create file
    canvas.toBlob((blob) => {
      if (!blob) return;

      // Create a File object from the blob
      const fileName = `camera_${Date.now()}.jpg`;
      const file = new File([blob], fileName, { type: 'image/jpeg' });

      // Store file for upload
      setSelectedFiles(prev => [...prev, file]);

      // Create preview URL
      const dataUrl = canvas.toDataURL('image/jpeg', 0.9);
      setSelectedImages(prev => [...prev, dataUrl]);

      // Close camera after capture
      stopCamera();
    }, 'image/jpeg', 0.9);
  }, [selectedImages.length, stopCamera]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    setUploadProgress('');

    try {
      let imageUrls: string[] = [];

      // Upload new images to MinIO if there are any files selected
      if (selectedFiles.length > 0) {
        setUploadProgress('Uploading images...');
        try {
          imageUrls = await uploadApi.uploadMultipleImages(selectedFiles, 'reviews');
        } catch (uploadErr) {
          console.error('Failed to upload images to MinIO, trying base64 fallback:', uploadErr);
          // Fallback: Use base64 data URLs if MinIO upload fails
          imageUrls = selectedImages.filter(img => img.startsWith('data:'));
        }
      } else if (selectedImages.length > 0) {
        // If editing and no new files, keep existing MinIO URLs
        imageUrls = selectedImages.filter(img => img.startsWith('http'));
      }

      setUploadProgress('');

      const submitData = {
        ...formData,
        images: imageUrls.length > 0 ? imageUrls : undefined,
      };

      if (editMode && userReview) {
        await reviewsApi.updateReview(userReview.id, submitData);
      } else {
        await reviewsApi.createReview(productId, submitData);
      }
      await loadReviews();
      await loadUserReview();
      setShowForm(false);
      setEditMode(false);
      setSelectedImages([]);
      setSelectedFiles([]);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to submit review');
    } finally {
      setSubmitting(false);
      setUploadProgress('');
    }
  };

  const handleDelete = async () => {
    if (!userReview) return;
    if (!confirm('Are you sure you want to delete your review?')) return;

    try {
      await reviewsApi.deleteReview(userReview.id);
      setUserReview(null);
      setFormData({ rating: 5, comment: '' });
      setSelectedImages([]);
      setSelectedFiles([]);
      await loadReviews();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to delete review');
    }
  };

  const handleEdit = () => {
    if (userReview) {
      setFormData({
        rating: userReview.rating,
        comment: userReview.comment,
      });
      // Load existing images for editing (MinIO URLs)
      if (userReview.images && userReview.images.length > 0) {
        setSelectedImages(userReview.images.map(img => img.url));
      } else {
        setSelectedImages([]);
      }
      // Reset selected files since we're editing existing review
      setSelectedFiles([]);
      setEditMode(true);
      setShowForm(true);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="w-6 h-6 text-[var(--color-primary)] animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold flex items-center gap-2">
            <MessageSquare className="w-5 h-5" />
            Reviews
          </h3>
          {reviewData && (
            <div className="flex items-center gap-2 mt-1">
              <StarRating rating={Math.round(reviewData.average_rating)} size="sm" />
              <span className="text-sm text-gray-600">
                {reviewData.average_rating.toFixed(1)} ({reviewData.total_reviews} reviews)
              </span>
            </div>
          )}
        </div>

        {isAuthenticated && !userReview && !showForm && (
          <button
            onClick={() => setShowForm(true)}
            className="px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-hover)] transition-colors"
          >
            Write Review
          </button>
        )}
      </div>

      {/* User's existing review */}
      {userReview && !showForm && (
        <div className="bg-[var(--color-primary-light)] border-2 border-[var(--color-primary)] rounded-xl p-4">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-sm font-medium text-[var(--color-primary)]">Your Review</p>
              <StarRating rating={userReview.rating} size="sm" />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleEdit}
                className="p-2 text-gray-500 hover:text-blue-600 hover:bg-blue-50 rounded-lg"
              >
                <Edit2 className="w-4 h-4" />
              </button>
              <button
                onClick={handleDelete}
                className="p-2 text-gray-500 hover:text-red-600 hover:bg-red-50 rounded-lg"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            </div>
          </div>
          {userReview.comment && (
            <p className="mt-2 text-gray-700">{userReview.comment}</p>
          )}
          {/* Display review images */}
          {userReview.images && userReview.images.length > 0 && (
            <div className="flex flex-wrap gap-2 mt-3">
              {userReview.images.map((img) => (
                <img
                  key={img.id}
                  src={img.url}
                  alt="Review"
                  className="w-16 h-16 object-cover rounded-lg border border-gray-200 cursor-pointer hover:opacity-80"
                  onClick={() => window.open(img.url, '_blank')}
                />
              ))}
            </div>
          )}
          <p className="text-xs text-gray-500 mt-2">{formatDate(userReview.created_at)}</p>
        </div>
      )}

      {/* Review Form */}
      {showForm && isAuthenticated && (
        <form onSubmit={handleSubmit} className="bg-gray-50 rounded-xl p-4 space-y-4">
          <div className="flex items-center justify-between">
            <h4 className="font-medium">{editMode ? 'Edit Review' : 'Write a Review'}</h4>
            <button
              type="button"
              onClick={() => { setShowForm(false); setEditMode(false); }}
              className="p-1 hover:bg-gray-200 rounded"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Your Rating</label>
            <StarRating
              rating={formData.rating}
              interactive
              onRate={(r) => setFormData({ ...formData, rating: r })}
              size="lg"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Your Review (optional)</label>
            <textarea
              value={formData.comment}
              onChange={(e) => setFormData({ ...formData, comment: e.target.value })}
              rows={4}
              className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)]"
              placeholder="Share your thoughts about this product..."
            />
          </div>

          {/* Image Upload Section */}
          <div>
            <label className="block text-sm font-medium mb-2">
              Add Photos (optional, max {MAX_IMAGES})
            </label>

            {/* Image Previews */}
            {selectedImages.length > 0 && (
              <div className="flex flex-wrap gap-2 mb-3">
                {selectedImages.map((img, index) => (
                  <div key={index} className="relative group">
                    <img
                      src={img}
                      alt={`Review image ${index + 1}`}
                      className="w-20 h-20 object-cover rounded-lg border border-gray-200"
                    />
                    <button
                      type="button"
                      onClick={() => removeImage(index)}
                      className="absolute -top-2 -right-2 bg-red-500 text-white rounded-full p-0.5 opacity-0 group-hover:opacity-100 transition-opacity"
                    >
                      <XCircle className="w-4 h-4" />
                    </button>
                  </div>
                ))}
              </div>
            )}

            {/* Upload Button */}
            {selectedImages.length < MAX_IMAGES && (
              <div>
                {/* Hidden file input for gallery selection */}
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/*"
                  multiple
                  onChange={handleImageSelect}
                  className="hidden"
                />
                {/* Hidden input for mobile camera capture (uses native camera app) */}
                <input
                  ref={cameraInputRef}
                  type="file"
                  accept="image/*"
                  capture="environment"
                  onChange={handleCameraCapture}
                  className="hidden"
                />
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={() => fileInputRef.current?.click()}
                    className="flex items-center gap-2 px-4 py-2 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-[var(--color-primary)] hover:text-[var(--color-primary)] transition-colors"
                  >
                    <ImagePlus className="w-5 h-5" />
                    <span>Add Photos</span>
                  </button>
                  <button
                    type="button"
                    onClick={handleTakePhoto}
                    className="flex items-center gap-2 px-4 py-2 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-[var(--color-primary)] hover:text-[var(--color-primary)] transition-colors"
                  >
                    <Camera className="w-5 h-5" />
                    <span>Take Photo</span>
                  </button>
                </div>
                <p className="text-xs text-gray-500 mt-1">
                  {selectedImages.length}/{MAX_IMAGES} photos
                </p>
              </div>
            )}

          </div>

          {error && (
            <p className="text-sm text-red-600">{error}</p>
          )}

          {uploadProgress && (
            <p className="text-sm text-blue-600 flex items-center gap-2">
              <Loader2 className="w-4 h-4 animate-spin" />
              {uploadProgress}
            </p>
          )}

          <button
            type="submit"
            disabled={submitting}
            className="w-full py-3 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-hover)] transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
          >
            {submitting && <Loader2 className="w-4 h-4 animate-spin" />}
            {editMode ? 'Update Review' : 'Submit Review'}
          </button>
        </form>
      )}

      {/* Reviews List */}
      <div className="space-y-4">
        {reviewData?.reviews && reviewData.reviews.length > 0 ? (
          reviewData.reviews
            .filter(review => review.id !== userReview?.id)
            .map((review) => (
              <div key={review.id} className="border-b border-gray-100 pb-4 last:border-0">
                <div className="flex items-start gap-3">
                  <div className="w-10 h-10 rounded-full bg-gradient-to-br from-[var(--color-primary)] to-pink-400 flex items-center justify-center text-white font-medium">
                    {review.user.name.charAt(0).toUpperCase()}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{review.user.name}</span>
                      <StarRating rating={review.rating} size="sm" />
                    </div>
                    {review.comment && (
                      <p className="mt-1 text-gray-700">{review.comment}</p>
                    )}
                    {/* Display review images */}
                    {review.images && review.images.length > 0 && (
                      <div className="flex flex-wrap gap-2 mt-2">
                        {review.images.map((img) => (
                          <img
                            key={img.id}
                            src={img.url}
                            alt="Review"
                            className="w-14 h-14 object-cover rounded-lg border border-gray-200 cursor-pointer hover:opacity-80"
                            onClick={() => window.open(img.url, '_blank')}
                          />
                        ))}
                      </div>
                    )}
                    <p className="text-xs text-gray-500 mt-1">{formatDate(review.created_at)}</p>
                  </div>
                </div>
              </div>
            ))
        ) : (
          !userReview && (
            <div className="text-center py-8 text-gray-500">
              <MessageSquare className="w-12 h-12 mx-auto mb-2 text-gray-300" />
              <p>No reviews yet. Be the first to review!</p>
            </div>
          )
        )}
      </div>

      {/* Camera Modal - Outside all other elements for proper rendering */}
      {showCamera && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            zIndex: 99999,
            backgroundColor: '#000',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          {/* Camera Header */}
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              padding: '16px',
              backgroundColor: 'rgba(0,0,0,0.7)',
            }}
          >
            <button
              type="button"
              onClick={stopCamera}
              style={{
                padding: '8px',
                color: 'white',
                background: 'transparent',
                border: 'none',
                borderRadius: '50%',
                cursor: 'pointer',
              }}
            >
              <X className="w-6 h-6" />
            </button>
            <span style={{ color: 'white', fontWeight: 500 }}>Take Photo</span>
            <button
              type="button"
              onClick={switchCamera}
              style={{
                padding: '8px',
                color: 'white',
                background: 'transparent',
                border: 'none',
                borderRadius: '50%',
                cursor: 'pointer',
              }}
            >
              <RefreshCw className="w-6 h-6" />
            </button>
          </div>

          {/* Video Preview - Use explicit flex:1 and min-height */}
          <div
            style={{
              flex: 1,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              backgroundColor: '#333',
              minHeight: '300px',
              overflow: 'hidden',
            }}
          >
            <video
              ref={videoRef}
              autoPlay
              playsInline
              muted
              style={{
                width: '100%',
                height: '100%',
                objectFit: 'cover',
                display: 'block',
              }}
            />
          </div>

          {/* Capture Button */}
          <div
            style={{
              display: 'flex',
              justifyContent: 'center',
              padding: '24px',
              backgroundColor: 'rgba(0,0,0,0.7)',
            }}
          >
            <button
              type="button"
              onClick={capturePhoto}
              style={{
                width: '64px',
                height: '64px',
                borderRadius: '50%',
                backgroundColor: 'white',
                border: '4px solid #ccc',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <div
                style={{
                  width: '48px',
                  height: '48px',
                  borderRadius: '50%',
                  backgroundColor: 'white',
                  border: '2px solid #999',
                }}
              />
            </button>
          </div>

          {/* Hidden canvas for capturing */}
          <canvas ref={canvasRef} style={{ display: 'none' }} />
        </div>
      )}
    </div>
  );
}
