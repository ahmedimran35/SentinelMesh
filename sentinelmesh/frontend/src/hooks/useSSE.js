import { useEffect, useRef, useCallback } from 'react';

export function useSSE(onEvent) {
  const sourceRef = useRef(null);

  useEffect(() => {
    const source = new EventSource('/events');
    sourceRef.current = source;

    const eventTypes = [
      'agent-status',
      'finding',
      'investigation-status',
      'commander-message',
      'alert',
      'report-ready',
    ];

    eventTypes.forEach(type => {
      source.addEventListener(type, (e) => {
        try {
          const data = JSON.parse(e.data);
          onEvent(type, data);
        } catch (err) {
          console.error('SSE parse error:', err);
        }
      });
    });

    source.onerror = () => {
      console.log('SSE connection error, reconnecting...');
    };

    return () => {
      source.close();
    };
  }, [onEvent]);

  return sourceRef;
}
