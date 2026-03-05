import React, { useState, useEffect } from 'react';
import { Loader2, MessageSquare, Trash2, Edit2, X } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { reviewsApi, Review, ReviewListResponse } from '../api/reviews';
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
      }
    } catch {
      setUserReview(null);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);

    try {
      if (editMode && userReview) {
        await reviewsApi.updateReview(userReview.id, formData);
      } else {
        await reviewsApi.createReview(productId, formData);
      }
      await loadReviews();
      await loadUserReview();
      setShowForm(false);
      setEditMode(false);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to submit review');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!userReview) return;
    if (!confirm('Are you sure you want to delete your review?')) return;

    try {
      await reviewsApi.deleteReview(userReview.id);
      setUserReview(null);
      setFormData({ rating: 5, comment: '' });
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

          {error && (
            <p className="text-sm text-red-600">{error}</p>
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
    </div>
  );
}
