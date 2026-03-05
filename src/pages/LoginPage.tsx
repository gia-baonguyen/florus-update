import React, { useState, useEffect } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { Eye, EyeOff, Mail, Lock, Sparkles } from 'lucide-react';
import { GoogleSignInButton } from '../components/GoogleSignInButton';

// Image URL from public folder - Vite serves files from public/ at root
const heroImageUrl = '/photo-1764694071531-008332b61f43.jpg';

export const LoginPage: React.FC = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [registerSuccess, setRegisterSuccess] = useState(false);
  const [resetSuccess, setResetSuccess] = useState(false);

  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    if (params.get('registered') === '1') {
      setRegisterSuccess(true);
      // Xoá query param khỏi URL cho sạch (không bắt buộc)
      const url = new URL(window.location.href);
      url.searchParams.delete('registered');
      window.history.replaceState({}, '', url.toString());
    }
    if (params.get('reset') === '1') {
      setResetSuccess(true);
      const url = new URL(window.location.href);
      url.searchParams.delete('reset');
      window.history.replaceState({}, '', url.toString());
    }
  }, [location.search]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      await login(email, password);
      navigate('/');
    } catch (err: any) {
      setError(err.response?.data?.message || 'Invalid email or password');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4" style={{ background: 'linear-gradient(135deg, #FFF8F8 0%, #ffffff 50%, #FFF0F0 100%)' }}>
      <div className="w-full max-w-[1200px] grid grid-cols-1 md:grid-cols-2 gap-12 items-center" style={{ gap: '3rem' }}>
        
        {/* Hero Section */}
        <div className="block w-full">
          <div className="relative rounded-2xl overflow-hidden shadow-[0_20px_40px_rgba(0,0,0,0.1)] bg-gray-200" style={{ minHeight: '600px' }}>
            <img 
              src={heroImageUrl}
              alt="Luxury Beauty Products"
              className="w-full h-full object-cover"
              style={{ width: '100%', height: '600px', objectFit: 'cover', display: 'block' }}
              onError={(e) => {
                console.error('❌ Image failed to load from:', heroImageUrl);
                console.error('Full URL attempted:', window.location.origin + heroImageUrl);
                console.error('Checking if file exists in public folder...');
                const target = e.target as HTMLImageElement;
                // Try fallback
                target.src = 'https://images.unsplash.com/photo-1764694071531-008332b61f43?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&q=80&w=1080';
              }}
              onLoad={(e) => {
                const img = e.target as HTMLImageElement;
                console.log('✅ Hero image loaded successfully');
                console.log('📐 Dimensions:', img.naturalWidth, 'x', img.naturalHeight);
                console.log('🔗 Source:', img.src);
              }}
            />
            <div className="absolute inset-0" style={{ background: 'linear-gradient(to top, rgba(0,0,0,0.6), rgba(0,0,0,0.2), transparent)' }}></div>
            <div className="absolute bottom-0 left-0 right-0 text-white" style={{ padding: '2rem' }}>
              <div className="flex items-center gap-2 mb-4">
                <Sparkles className="w-6 h-6 text-[#FF6B6B]" />
                <span className="font-serif" style={{ fontSize: '1.5rem' }}>Florus Beauty</span>
              </div>
              <h2 className="font-serif font-medium" style={{ fontSize: '2rem', marginBottom: '0.75rem' }}>
                Intelligent Glow
              </h2>
              <p className="leading-relaxed" style={{ fontSize: '0.875rem', lineHeight: '1.6', maxWidth: '400px', color: 'rgba(255,255,255,0.9)' }}>
                Experience personalized beauty with AI-powered recommendations. 
                Your journey to radiant skin starts here.
              </p>
            </div>
          </div>
        </div>

        {/* Form Section */}
        <div className="bg-white rounded-2xl border border-gray-100 px-6 py-10 md:px-10 md:py-12 shadow-[0_10px_40px_rgba(0,0,0,0.08)]">
          
          {/* Mobile Logo */}
          <div className="md:hidden flex items-center gap-2 mb-8 text-gray-800">
            <Sparkles className="w-6 h-6 text-[#FF6B6B]" />
            <span className="font-serif text-2xl">Florus Beauty</span>
          </div>

          {/* Header */}
          <div className="mb-8">
            <h1 className="font-serif text-gray-800 mb-2 font-medium" style={{ fontSize: '2rem' }}>
              <span className="md:hidden">Welcome back</span>
              <span className="hidden md:inline" style={{ fontSize: '2.5rem' }}>Welcome back</span>
            </h1>
            <p className="text-gray-500" style={{ fontSize: '0.875rem', color: '#666' }}>
              Sign in to continue your beauty journey
            </p>
          </div>

          {/* Register success message */}
          {registerSuccess && (
            <div className="bg-green-50 text-green-700 px-4 py-3 rounded-xl text-sm border border-green-100 mb-4">
              Account created successfully. Please sign in with your new credentials.
            </div>
          )}

          {/* Reset password success message */}
          {resetSuccess && (
            <div className="bg-green-50 text-green-700 px-4 py-3 rounded-xl text-sm border border-green-100 mb-4">
              Password has been reset successfully. Please sign in with your new password.
            </div>
          )}

          {/* Error Message */}
          {error && (
            <div className="bg-red-50 text-red-600 px-4 py-3 rounded-xl text-sm border border-red-100 mb-5">
              {error}
            </div>
          )}

          {/* Login Form */}
          <form onSubmit={handleSubmit} className="mb-6">
            {/* Email */}
            <div className="mb-5">
              <label htmlFor="email" className="block font-medium mb-2" style={{ fontSize: '0.875rem', color: '#374151' }}>
                Email
              </label>
              <div className="relative">
                <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5" style={{ color: '#9CA3AF' }} />
                <input
                  type="email"
                  id="email"
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
                    color: '#333'
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
            <div className="mb-5">
              <label htmlFor="password" className="block font-medium mb-2" style={{ fontSize: '0.875rem', color: '#374151' }}>
                Password
              </label>
              <div className="relative">
                <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5" style={{ color: '#9CA3AF' }} />
                <input
                  type={showPassword ? 'text' : 'password'}
                  id="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Enter password"
                  autoComplete="current-password"
                  className="w-full border bg-white text-gray-800 transition-all focus:outline-none placeholder:text-[#9CA3AF]"
                  style={{
                    padding: '0.875rem 3rem 0.875rem 3rem',
                    borderColor: '#E5E7EB',
                    borderRadius: '0.75rem',
                    fontSize: '0.9375rem',
                    color: '#333'
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
                  className="absolute right-4 top-1/2 -translate-y-1/2 transition-colors p-1"
                  style={{ color: '#9CA3AF' }}
                  onMouseEnter={(e) => e.currentTarget.style.color = '#4B5563'}
                  onMouseLeave={(e) => e.currentTarget.style.color = '#9CA3AF'}
                >
                  {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
            </div>

            {/* Remember & Forgot */}
            <div className="flex items-center justify-between text-sm mb-5">
              <label className="flex items-center gap-2 cursor-pointer text-gray-600">
                <input 
                  type="checkbox" 
                  className="w-4 h-4 rounded border-gray-300 text-[#FF6B6B] focus:ring-[#FF6B6B] cursor-pointer"
                />
                <span>Remember me</span>
              </label>
              <button 
                type="button"
                className="text-[#FF6B6B] hover:text-[#E85555] font-medium transition-colors text-sm"
                onClick={() => navigate('/forgot-password')}
              >
                Forgot password?
              </button>
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              disabled={isLoading}
              className="w-full text-white border-none cursor-pointer transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              style={{
                padding: '0.875rem',
                background: '#FF6B6B',
                borderRadius: '0.75rem',
                fontSize: '0.9375rem',
                fontWeight: 500,
                boxShadow: '0 4px 12px rgba(255, 107, 107, 0.3)'
              }}
              onMouseEnter={(e) => {
                if (!isLoading) {
                  e.currentTarget.style.background = '#E85555';
                  e.currentTarget.style.transform = 'translateY(-1px)';
                  e.currentTarget.style.boxShadow = '0 6px 16px rgba(255, 107, 107, 0.4)';
                }
              }}
              onMouseLeave={(e) => {
                if (!isLoading) {
                  e.currentTarget.style.background = '#FF6B6B';
                  e.currentTarget.style.transform = 'translateY(0)';
                  e.currentTarget.style.boxShadow = '0 4px 12px rgba(255, 107, 107, 0.3)';
                }
              }}
            >
              {isLoading ? 'Signing in...' : 'Sign In'}
            </button>
          </form>

          {/* Divider */}
          <div className="flex items-center" style={{ gap: '1rem', margin: '1.5rem 0' }}>
            <div className="flex-1" style={{ height: '1px', background: '#E5E7EB' }}></div>
            <span style={{ fontSize: '0.75rem', color: '#9CA3AF', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Or</span>
            <div className="flex-1" style={{ height: '1px', background: '#E5E7EB' }}></div>
          </div>

          {/* Google Sign In */}
          <div className="mb-6">
            <GoogleSignInButton
              onSuccess={() => navigate('/')}
              onError={(err) => setError(err)}
            />
          </div>

          {/* Sign Up Link */}
          <p className="text-center text-gray-600 text-sm mt-6">
            Don't have an account?{' '}
            <Link to="/register" className="text-[#FF6B6B] font-medium hover:text-[#E85555] transition-colors">
              Sign Up
            </Link>
          </p>

          {/* Demo Hint */}
          <div className="mt-6 p-4 bg-[#D4F4EB] border border-[#1FAB89]/20 rounded-xl text-center text-xs text-gray-700">
            <p className="mb-1">
              <span className="font-semibold">Demo accounts:</span>
            </p>
            <p>Admin: admin@florus.com / admin123</p>
            <p className="mt-1">
              Or create a new account on the <Link to="/register" className="text-[#FF6B6B] font-medium hover:text-[#E85555]">Sign Up</Link> page.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default LoginPage;