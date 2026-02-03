interface StatusBarProps {
  isConnected: boolean;
  lastSync: Date | null;
  projectStatus?: string;
}

export function StatusBar({ isConnected, lastSync, projectStatus }: StatusBarProps) {
  const formatTime = (date: Date) => {
    return date.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  return (
    <footer className="flex items-center justify-between px-4 py-1 text-xs border-t border-vscode-border bg-vscode-bg opacity-80">
      <div className="flex items-center gap-4">
        {/* Connection Status */}
        <div className="flex items-center gap-1">
          <span
            className={`w-2 h-2 rounded-full ${
              isConnected ? 'bg-green-500' : 'bg-red-500'
            }`}
          />
          <span>{isConnected ? 'Connected' : 'Disconnected'}</span>
        </div>

        {/* Last Sync */}
        {lastSync && (
          <div className="opacity-70">
            Last sync: {formatTime(lastSync)}
          </div>
        )}
      </div>

      <div className="flex items-center gap-4">
        {/* Project Status */}
        {projectStatus && (
          <div className="flex items-center gap-1">
            <span className="opacity-70">Status:</span>
            <span className={getStatusColor(projectStatus)}>{projectStatus}</span>
          </div>
        )}
      </div>
    </footer>
  );
}

function getStatusColor(status: string): string {
  switch (status.toLowerCase()) {
    case 'active':
      return 'text-green-400';
    case 'pending':
      return 'text-yellow-400';
    case 'completed':
      return 'text-blue-400';
    case 'archived':
      return 'text-gray-400';
    default:
      return '';
  }
}
