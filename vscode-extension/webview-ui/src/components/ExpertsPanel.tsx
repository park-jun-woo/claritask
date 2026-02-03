import React, { useState } from 'react';
import { useStore } from '../store';
import ExpertCard from './ExpertCard';
import CreateExpertDialog from './CreateExpertDialog';

const ExpertsPanel: React.FC = () => {
  const { experts, selectedExpertId, setSelectedExpert } = useStore();
  const [showCreateDialog, setShowCreateDialog] = useState(false);

  const assignedExperts = experts.filter((e) => e.assigned);
  const availableExperts = experts.filter((e) => !e.assigned);

  return (
    <div className="flex h-full">
      {/* Left: Expert list */}
      <div className="w-1/3 border-r border-vscode-border overflow-y-auto p-4">
        <h3 className="font-bold mb-2 text-vscode-foreground">Assigned</h3>
        {assignedExperts.length === 0 ? (
          <p className="text-sm opacity-60 mb-4">No experts assigned</p>
        ) : (
          assignedExperts.map((expert) => (
            <div
              key={expert.id}
              className={`p-2 cursor-pointer rounded mb-1 ${
                selectedExpertId === expert.id
                  ? 'bg-vscode-list-active-bg text-vscode-list-active-fg'
                  : 'hover:bg-vscode-list-hover'
              }`}
              onClick={() => setSelectedExpert(expert.id)}
            >
              <div className="font-medium">{expert.name}</div>
              <div className="text-sm opacity-70">{expert.domain}</div>
            </div>
          ))
        )}

        <h3 className="font-bold mb-2 mt-4 text-vscode-foreground">Available</h3>
        {availableExperts.length === 0 ? (
          <p className="text-sm opacity-60 mb-4">No available experts</p>
        ) : (
          availableExperts.map((expert) => (
            <div
              key={expert.id}
              className={`p-2 cursor-pointer rounded mb-1 opacity-60 ${
                selectedExpertId === expert.id
                  ? 'bg-vscode-list-active-bg text-vscode-list-active-fg'
                  : 'hover:bg-vscode-list-hover'
              }`}
              onClick={() => setSelectedExpert(expert.id)}
            >
              <div className="font-medium">{expert.name}</div>
              <div className="text-sm opacity-70">{expert.domain}</div>
            </div>
          ))
        )}

        <button
          onClick={() => setShowCreateDialog(true)}
          className="mt-4 w-full p-2 bg-vscode-button-bg text-vscode-button-fg rounded hover:bg-vscode-button-hover-bg"
        >
          + Create New Expert
        </button>
      </div>

      {/* Right: Expert detail */}
      <div className="w-2/3 p-4 overflow-y-auto">
        {selectedExpertId ? (
          <ExpertCard expertId={selectedExpertId} />
        ) : (
          <div className="text-vscode-foreground opacity-60 text-center mt-10">
            Select an expert to view details
          </div>
        )}
      </div>

      <CreateExpertDialog
        isOpen={showCreateDialog}
        onClose={() => setShowCreateDialog(false)}
      />
    </div>
  );
};

export default ExpertsPanel;
