import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Check, X, Loader2, ShoppingBag, ArrowLeft } from 'lucide-react';
import { paymentsApi } from '../api/payments';

type PaymentStatus = 'loading' | 'success' | 'failed' | 'cancelled';

export function PaymentResultPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [status, setStatus] = useState<PaymentStatus>('loading');
  const [orderCode, setOrderCode] = useState<string>('');
  const [errorMessage, setErrorMessage] = useState<string>('');

  useEffect(() => {
    const checkPaymentStatus = async () => {
      // Get order code from URL params (different payment gateways use different param names)
      const code = searchParams.get('order_code') ||
                   searchParams.get('vnp_TxnRef') ||
                   searchParams.get('orderId') ||
                   searchParams.get('apptransid')?.split('_')[1] || '';

      if (!code) {
        setStatus('failed');
        setErrorMessage('Order code not found');
        return;
      }

      setOrderCode(code);

      // Check for immediate status from URL params
      const vnpResponseCode = searchParams.get('vnp_ResponseCode');
      const momoResultCode = searchParams.get('resultCode');
      const zaloPayStatus = searchParams.get('status');

      // VNPay: 00 = success
      if (vnpResponseCode) {
        if (vnpResponseCode === '00') {
          setStatus('success');
        } else if (vnpResponseCode === '24') {
          setStatus('cancelled');
        } else {
          setStatus('failed');
          setErrorMessage('Payment was not successful');
        }
        return;
      }

      // MoMo: 0 = success
      if (momoResultCode !== null) {
        if (momoResultCode === '0') {
          setStatus('success');
        } else if (momoResultCode === '1006') {
          setStatus('cancelled');
        } else {
          setStatus('failed');
          setErrorMessage('Payment was not successful');
        }
        return;
      }

      // ZaloPay: 1 = success
      if (zaloPayStatus) {
        if (zaloPayStatus === '1') {
          setStatus('success');
        } else if (zaloPayStatus === '-49') {
          setStatus('cancelled');
        } else {
          setStatus('failed');
          setErrorMessage('Payment was not successful');
        }
        return;
      }

      // If no immediate status, check with backend
      try {
        const paymentStatus = await paymentsApi.getPaymentStatus(code);
        if (paymentStatus.status === 'success') {
          setStatus('success');
        } else if (paymentStatus.status === 'cancelled') {
          setStatus('cancelled');
        } else if (paymentStatus.status === 'failed') {
          setStatus('failed');
          setErrorMessage(paymentStatus.error_message || 'Payment failed');
        } else {
          // Still pending - show as loading and check again
          setStatus('loading');
        }
      } catch {
        setStatus('failed');
        setErrorMessage('Unable to verify payment status');
      }
    };

    checkPaymentStatus();
  }, [searchParams]);

  const getStatusContent = () => {
    switch (status) {
      case 'loading':
        return {
          icon: <Loader2 className="w-12 h-12 text-[var(--color-primary)] animate-spin" />,
          bgColor: 'bg-[var(--color-primary-light)]',
          title: 'Processing Payment',
          titleColor: 'text-[var(--color-primary)]',
          description: 'Please wait while we verify your payment...',
        };
      case 'success':
        return {
          icon: <Check className="w-12 h-12 text-white" />,
          bgColor: 'bg-green-500',
          title: 'Payment Successful!',
          titleColor: 'text-green-700',
          description: 'Thank you for your purchase. Your order has been confirmed.',
        };
      case 'cancelled':
        return {
          icon: <X className="w-12 h-12 text-white" />,
          bgColor: 'bg-amber-500',
          title: 'Payment Cancelled',
          titleColor: 'text-amber-700',
          description: 'You cancelled the payment. Your order is still pending.',
        };
      case 'failed':
        return {
          icon: <X className="w-12 h-12 text-white" />,
          bgColor: 'bg-red-500',
          title: 'Payment Failed',
          titleColor: 'text-red-700',
          description: errorMessage || 'There was an issue processing your payment.',
        };
    }
  };

  const content = getStatusContent();

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      <div className="container py-12">
        {/* Back button */}
        <button
          onClick={() => navigate('/')}
          className="flex items-center gap-2 text-[var(--color-text-secondary)] hover:text-[var(--color-primary)] transition-colors mb-6"
        >
          <ArrowLeft className="w-5 h-5" />
          Back to Home
        </button>

        <div className="max-w-md mx-auto text-center bg-white rounded-2xl p-12 shadow-[var(--shadow-md)] border border-[var(--color-border)]">
          <div className={`w-24 h-24 mx-auto mb-6 ${content.bgColor} rounded-full flex items-center justify-center`}>
            {content.icon}
          </div>

          <h2 className={`text-2xl font-serif mb-3 ${content.titleColor}`}>
            {content.title}
          </h2>

          <p className="text-[var(--color-text-secondary)] mb-4">
            {content.description}
          </p>

          {orderCode && (
            <p className="text-lg font-medium mb-6">
              Order Code: <span className="text-[var(--color-primary)]">{orderCode}</span>
            </p>
          )}

          <div className="flex flex-col gap-3">
            {status === 'success' && (
              <>
                <button
                  onClick={() => navigate('/orders')}
                  className="w-full px-6 py-3 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-dark)] transition-colors flex items-center justify-center gap-2"
                >
                  <ShoppingBag className="w-5 h-5" />
                  View My Orders
                </button>
                <button
                  onClick={() => navigate('/')}
                  className="w-full px-6 py-3 border-2 border-[var(--color-border)] rounded-full hover:border-[var(--color-primary)] transition-colors"
                >
                  Continue Shopping
                </button>
              </>
            )}

            {status === 'cancelled' && (
              <>
                <button
                  onClick={() => navigate('/orders')}
                  className="w-full px-6 py-3 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-dark)] transition-colors"
                >
                  View Order & Retry Payment
                </button>
                <button
                  onClick={() => navigate('/')}
                  className="w-full px-6 py-3 border-2 border-[var(--color-border)] rounded-full hover:border-[var(--color-primary)] transition-colors"
                >
                  Continue Shopping
                </button>
              </>
            )}

            {status === 'failed' && (
              <>
                <button
                  onClick={() => navigate('/orders')}
                  className="w-full px-6 py-3 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-dark)] transition-colors"
                >
                  View Order & Try Again
                </button>
                <button
                  onClick={() => navigate('/')}
                  className="w-full px-6 py-3 border-2 border-[var(--color-border)] rounded-full hover:border-[var(--color-primary)] transition-colors"
                >
                  Return to Home
                </button>
              </>
            )}

            {status === 'loading' && (
              <p className="text-sm text-[var(--color-text-muted)]">
                This may take a few moments...
              </p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export default PaymentResultPage;
