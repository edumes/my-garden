import { Button } from '@/components/ui/button';
import { apiService } from '@/services/api';
import { CheckCircle, Loader, XCircle } from 'lucide-react';
import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

const JoinGardenPage: React.FC = () => {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [message, setMessage] = useState('');

  useEffect(() => {
    const joinGarden = async () => {
      console.log('JoinGardenPage: token =', token);

      if (!token) {
        setStatus('error');
        setMessage('Invalid access token');
        return;
      }

      try {
        console.log('JoinGardenPage: Attempting to join garden with token:', token);
        const response = await apiService.joinGarden({ token });
        console.log('JoinGardenPage: Success response:', response);
        setStatus('success');
        setMessage(`Successfully joined garden: ${response.garden.name}`);

        setTimeout(() => {
          navigate(`/garden/${response.garden.id}`);
        }, 2000);
      } catch (error: any) {
        console.error('JoinGardenPage: Error joining garden:', error);
        setStatus('error');
        setMessage(error.message || 'Failed to join garden');
      }
    };

    joinGarden();
  }, [token, navigate]);

  const renderContent = () => {
    switch (status) {
      case 'loading':
        return (
          <div className="text-center">
            <Loader className="w-8 h-8 text-green-600 animate-spin mx-auto mb-4" />
            <p className="text-lg font-medium">Joining garden...</p>
            <p className="text-muted-foreground">Please wait while we process your request</p>
          </div>
        );

      case 'success':
        return (
          <div className="text-center">
            <CheckCircle className="w-12 h-12 text-green-600 mx-auto mb-4" />
            <p className="text-lg font-medium text-green-800">{message}</p>
            <p className="text-muted-foreground">Redirecting to garden...</p>
          </div>
        );

      case 'error':
        return (
          <div className="text-center">
            <XCircle className="w-12 h-12 text-red-600 mx-auto mb-4" />
            <p className="text-lg font-medium text-red-800">{message}</p>
            <div className="mt-4 space-x-2">
              <Button onClick={() => navigate('/')}>
                Go to Dashboard
              </Button>
              <Button variant="outline" onClick={() => window.location.reload()}>
                Try Again
              </Button>
            </div>
          </div>
        );
    }
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center transition-colors duration-200">
      <div className="max-w-md w-full mx-4">
        {renderContent()}
      </div>
    </div>
  );
};

export default JoinGardenPage; 