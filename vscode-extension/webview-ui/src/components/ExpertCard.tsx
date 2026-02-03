import React from 'react';
import { useStore } from '../store';
import { assignExpert, unassignExpert, openExpertFile } from '../vscode';

interface ExpertCardProps {
  expertId: string;
}

const ExpertCard: React.FC<ExpertCardProps> = ({ expertId }) => {
  const { getExpert } = useStore();
  const expert = getExpert(expertId);

  if (!expert) return null;

  const handleAssign = () => {
    if (expert.assigned) {
      unassignExpert(expert.id);
    } else {
      assignExpert(expert.id);
    }
  };

  const handleOpenFile = () => {
    openExpertFile(expert.id);
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex justify-between items-start">
        <div>
          <h2 className="text-xl font-bold text-vscode-foreground">{expert.name}</h2>
          <p className="text-vscode-foreground opacity-70">{expert.domain}</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={handleAssign}
            className={`px-3 py-1 rounded text-sm ${
              expert.assigned
                ? 'bg-red-700 text-white hover:bg-red-800'
                : 'bg-green-700 text-white hover:bg-green-800'
            }`}
          >
            {expert.assigned ? 'Unassign' : 'Assign'}
          </button>
          <button
            onClick={handleOpenFile}
            className="px-3 py-1 bg-vscode-button-bg text-vscode-button-fg rounded text-sm hover:bg-vscode-button-hover-bg"
          >
            Open File
          </button>
        </div>
      </div>

      {/* Meta info */}
      <div className="grid grid-cols-2 gap-2 text-sm border-t border-vscode-border pt-4">
        <div>
          <span className="opacity-60">Language:</span> {expert.language || '-'}
        </div>
        <div>
          <span className="opacity-60">Framework:</span> {expert.framework || '-'}
        </div>
        <div>
          <span className="opacity-60">Version:</span> {expert.version}
        </div>
        <div>
          <span className="opacity-60">Status:</span>{' '}
          <span
            className={`px-1 py-0.5 rounded text-xs ${
              expert.status === 'active' ? 'bg-green-700' : 'bg-gray-600'
            }`}
          >
            {expert.status}
          </span>
        </div>
      </div>

      {/* Content preview */}
      <div className="border-t border-vscode-border pt-4">
        <h3 className="font-bold mb-2 text-vscode-foreground">Content</h3>
        <div className="bg-vscode-editor-bg p-4 rounded overflow-auto max-h-96">
          <pre className="text-sm whitespace-pre-wrap font-mono">{expert.content}</pre>
        </div>
      </div>
    </div>
  );
};

export default ExpertCard;
