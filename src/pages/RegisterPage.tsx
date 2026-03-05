import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { Eye, EyeOff, Mail, Lock, User, Sparkles } from 'lucide-react';
import { GoogleSignInButton } from '../components/GoogleSignInButton';

// Dùng cùng ảnh hero với LoginPage (đặt trong public/)
const heroImageUrl = '/photo-1764694071531-008332b61f43.jpg';

export const RegisterPage: React.FC = () => {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const { register } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (password.length < 6) {
      setError('Password must be at least 6 characters');
      return;
    }

    setIsLoading(true);

    try {
      await register(name, email, password);
      // Đăng ký thành công: chuyển về trang đăng nhập với thông báo
      navigate('/login?registered=1');
    } catch (err: any) {
      setError(err.response?.data?.message || 'Registration failed. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-[#FFF5F5] via-[#FBD1DC] to-[#F8B4C6] flex items-center justify-center px-4 py-8">
      <div className="w-full max-w-5xl grid md:grid-cols-2 gap-8">
        {/* Visual side - đồng bộ style ảnh với LoginPage */}
        <div className="hidden md:flex flex-col justify-between rounded-3xl bg-gradient-to-br from-[#E87A90] via-[#E87A90] to-[#D4AF37] p-8 text-white shadow-2xl">
          <div className="space-y-6">
            <div className="inline-flex items-center gap-3 mb-2">
              <div className="w-12 h-12 bg-white/10 rounded-2xl flex items-center justify-center shadow-lg">
                <Sparkles className="w-6 h-6 text-white" />
              </div>
              <span className="text-3xl font-serif font-bold tracking-wide">Florus</span>
            </div>

            <div className="relative rounded-3xl overflow-hidden shadow-[0_18px_40px_rgba(0,0,0,0.35)]">
              <img
                src={heroImageUrl}
                alt="Florus Beauty Collection"
                className="w-full h-[320px] object-cover"
                onError={(e) => {
                  const target = e.target as HTMLImageElement;
                  target.src =
                    'https://images.unsplash.com/photo-1764694071531-008332b61f43?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&q=80&w=1080';
                }}
              />
              <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-black/25 to-transparent" />
              <div className="absolute bottom-0 left-0 right-0 p-5">
                <h2 className="text-2xl font-serif font-semibold mb-2">
                  Join the Florus collective
                </h2>
                <p className="text-xs text-white/90 max-w-xs leading-relaxed">
                  Create your account to sync wishlist, track orders, and unlock AI-powered beauty experiences.
                </p>
              </div>
            </div>
          </div>

          <div className="mt-6 text-xs text-white/80 space-y-1">
            <p className="font-medium">Curated for modern beauty lovers</p>
            <p>Fast onboarding • Secure • Delightfully simple</p>
          </div>
        </div>

        {/* Register Form */}
        <div className="bg-white rounded-3xl shadow-[0_10px_25px_rgba(0,0,0,0.05)] border border-[#F1E4E4] px-8 py-10 flex flex-col justify-center">
          {/* Heading */}
          <div className="mb-8">
            <div className="inline-flex items-center gap-3 mb-4">
              <div className="w-10 h-10 rounded-2xl bg-[#E87A90]/10 flex items-center justify-center">
                <Sparkles className="w-5 h-5 text-[#E87A90]" />
              </div>
              <span className="text-2xl font-serif font-semibold text-[#4A5568]">
                Create account
              </span>
            </div>
            <p className="mt-1 text-sm text-[#718096] max-w-sm">
              Start your Florus experience in just a few seconds. No spam, only beauty.
            </p>
          </div>

          {/* Form */}
          <div>
          <form onSubmit={handleSubmit} className="space-y-5">
            {error && (
              <div className="bg-red-50 text-red-600 px-4 py-3 rounded-xl text-sm border border-red-100">
                {error}
              </div>
            )}

            {/* Name */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Full Name
              </label>
              <div className="relative">
                <User className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="John Doe"
                  className="w-full border bg-white text-gray-800 transition-all focus:outline-none placeholder:text-[#9CA3AF]"
                  style={{
                    padding: '0.875rem 1rem 0.875rem 3rem',
                    borderColor: '#E5E7EB',
                    borderRadius: '0.75rem',
                    fontSize: '0.9375rem',
                    color: '#333',
                  }}
                  onFocus={(e) => {
                    e.target.style.borderColor = '#FF6B6B';
                    e.target.style.boxShadow = '0 0 0 3px rgba(255, 107, 107, 0.1)';
                  }}
                  onBlur={(e) => {
                    e.target.style.borderColor = '#E5E7EB';
                    e.target.style.boxShadow = 'none';
                  }}
                  required
                />
              </div>
            </div>

            {/* Email */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Email
              </label>
              <div className="relative">
                <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="your@email.com"
                  autoComplete="email"
                  className="w-full border bg-white text-gray-800 transition-all focus:outline-none placeholder:text-[#9CA3AF]"
                  style={{
                    padding: '0.875rem 1rem 0.875rem 3rem',
                    borderColor: '#E5E7EB',
                    borderRadius: '0.75rem',
                    fontSize: '0.9375rem',
                    color: '#333',
                  }}
                  onFocus={(e) => {
                    e.target.style.borderColor = '#FF6B6B';
                    e.target.style.boxShadow = '0 0 0 3px rgba(255, 107, 107, 0.1)';
                  }}
                  onBlur={(e) => {
                    e.target.style.borderColor = '#E5E7EB';
                    e.target.style.boxShadow = 'none';
                  }}
                  required
                />
              </div>
            </div>

            {/* Password */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Password
              </label>
              <div className="relative">
                <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Minimum 6 characters"
                  autoComplete="new-password"
                  className="w-full border bg-white text-gray-800 transition-all focus:outline-none placeholder:text-[#9CA3AF]"
                  style={{
                    padding: '0.875rem 3rem 0.875rem 3rem',
                    borderColor: '#E5E7EB',
                    borderRadius: '0.75rem',
                    fontSize: '0.9375rem',
                    color: '#333',
                  }}
                  onFocus={(e) => {
                    e.target.style.borderColor = '#FF6B6B';
                    e.target.style.boxShadow = '0 0 0 3px rgba(255, 107, 107, 0.1)';
                  }}
                  onBlur={(e) => {
                    e.target.style.borderColor = '#E5E7EB';
                    e.target.style.boxShadow = 'none';
                  }}
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                >
                  {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
            </div>

            {/* Confirm Password */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Confirm Password
              </label>
              <div className="relative">
                <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  placeholder="Re-enter password"
                  autoComplete="new-password"
                  className="w-full border bg-white text-gray-800 transition-all focus:outline-none placeholder:text-[#9CA3AF]"
                  style={{
                    padding: '0.875rem 1rem 0.875rem 3rem',
                    borderColor: '#E5E7EB',
                    borderRadius: '0.75rem',
                    fontSize: '0.9375rem',
                    color: '#333',
                  }}
                  onFocus={(e) => {
                    e.target.style.borderColor = '#FF6B6B';
                    e.target.style.boxShadow = '0 0 0 3px rgba(255, 107, 107, 0.1)';
                  }}
                  onBlur={(e) => {
                    e.target.style.borderColor = '#E5E7EB';
                    e.target.style.boxShadow = 'none';
                  }}
                  required
                />
              </div>
            </div>

            {/* Terms */}
            <div className="flex items-start gap-3">
              <input
                type="checkbox"
                id="terms"
                className="mt-1 w-4 h-4 rounded border-gray-300 text-[#E87A90] focus:ring-[#E87A90]"
                required
              />
              <label htmlFor="terms" className="text-sm text-gray-600">
                I agree to the{' '}
                <a href="#" className="text-[var(--color-primary)] hover:underline">
                  Terms of Service
                </a>{' '}
                and{' '}
                <a href="#" className="text-[var(--color-primary)] hover:underline">
                  Privacy Policy
                </a>
              </label>
            </div>

            {/* Submit */}
            <button
              type="submit"
              disabled={isLoading}
              className="w-full py-3.5 text-white rounded-xl font-medium tracking-wide transition-all disabled:opacity-50 mt-2"
              style={{
                background: '#E87A90',
                boxShadow: '0 10px 25px rgba(232,122,144,0.35)',
                border: 'none',
              }}
              onMouseEnter={(e) => {
                if (!isLoading) {
                  e.currentTarget.style.background = '#d66a81';
                }
              }}
              onMouseLeave={(e) => {
                if (!isLoading) {
                  e.currentTarget.style.background = '#E87A90';
                }
              }}
            >
              {isLoading ? 'Registering...' : 'Register'}
            </button>
          </form>

          {/* Divider */}
          <div className="flex items-center gap-4 my-6">
            <div className="flex-1 h-px bg-gray-200"></div>
            <span className="text-xs text-gray-400 uppercase tracking-wide">Or continue with</span>
            <div className="flex-1 h-px bg-gray-200"></div>
          </div>

          {/* Social Login */}
          <div className="space-y-3">
            <GoogleSignInButton
              onSuccess={() => navigate('/')}
              onError={(error) => setError(error)}
            />
          </div>

          {/* Login Link */}
          <p className="text-center text-[#718096] mt-6 text-sm">
            Already have an account?{' '}
            <Link to="/login" className="text-[#E87A90] font-medium hover:underline">
              Sign In
            </Link>
          </p>
        </div>
      </div>
    </div>
  </div>
  );
};

export default RegisterPage;
