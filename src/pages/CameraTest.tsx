import React, { useState, useRef, useEffect } from 'react';

export function CameraTest() {
  const [stream, setStream] = useState<MediaStream | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [status, setStatus] = useState<string>('Click "Start Camera" to begin');
  const videoRef = useRef<HTMLVideoElement>(null);

  const startCamera = async () => {
    setStatus('Requesting camera access...');
    setError(null);

    try {
      const mediaStream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: 'environment' },
        audio: false,
      });

      setStatus('Got stream, setting up video...');
      setStream(mediaStream);

      if (videoRef.current) {
        videoRef.current.srcObject = mediaStream;
        setStatus('srcObject set, playing...');

        try {
          await videoRef.current.play();
          setStatus('Camera is playing!');
        } catch (playErr) {
          setStatus(`Play failed: ${playErr}`);
        }
      } else {
        setStatus('videoRef.current is null!');
      }
    } catch (err: any) {
      setError(`Camera error: ${err.message}`);
      setStatus('Failed to access camera');
    }
  };

  const stopCamera = () => {
    if (stream) {
      stream.getTracks().forEach(track => track.stop());
      setStream(null);
    }
    setStatus('Camera stopped');
  };

  useEffect(() => {
    return () => {
      if (stream) {
        stream.getTracks().forEach(track => track.stop());
      }
    };
  }, [stream]);

  return (
    <div style={{ padding: '20px', maxWidth: '600px', margin: '0 auto' }}>
      <h1 style={{ fontSize: '24px', marginBottom: '20px' }}>Camera Test Page</h1>

      <div style={{ marginBottom: '20px' }}>
        <p><strong>Status:</strong> {status}</p>
        {error && <p style={{ color: 'red' }}><strong>Error:</strong> {error}</p>}
      </div>

      <div style={{ marginBottom: '20px' }}>
        <button
          onClick={startCamera}
          style={{
            padding: '10px 20px',
            marginRight: '10px',
            backgroundColor: '#4CAF50',
            color: 'white',
            border: 'none',
            borderRadius: '5px',
            cursor: 'pointer'
          }}
        >
          Start Camera
        </button>
        <button
          onClick={stopCamera}
          style={{
            padding: '10px 20px',
            backgroundColor: '#f44336',
            color: 'white',
            border: 'none',
            borderRadius: '5px',
            cursor: 'pointer'
          }}
        >
          Stop Camera
        </button>
      </div>

      <div style={{
        width: '100%',
        height: '400px',
        backgroundColor: '#333',
        border: '3px solid #666',
        borderRadius: '10px',
        overflow: 'hidden'
      }}>
        <video
          ref={videoRef}
          autoPlay
          playsInline
          muted
          style={{
            width: '100%',
            height: '100%',
            objectFit: 'cover',
            display: 'block'
          }}
        />
      </div>

      <div style={{ marginTop: '20px', fontSize: '14px', color: '#666' }}>
        <p><strong>Debug info:</strong></p>
        <p>videoRef exists: {videoRef.current ? 'Yes' : 'No'}</p>
        <p>Stream exists: {stream ? 'Yes' : 'No'}</p>
        <p>Stream tracks: {stream ? stream.getTracks().length : 0}</p>
        <p>Protocol: {window.location.protocol}</p>
        <p>Hostname: {window.location.hostname}</p>
      </div>
    </div>
  );
}

export default CameraTest;
