import React, { useState } from 'react';
import { createExpert } from '../vscode';

interface CreateExpertDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

const CreateExpertDialog: React.FC<CreateExpertDialogProps> = ({ isOpen, onClose }) => {
  const [expertId, setExpertId] = useState('');
  const [error, setError] = useState('');

  if (!isOpen) return null;

  const validateId = (id: string): boolean => {
    const pattern = /^[a-z0-9-]+$/;
    return pattern.test(id) && id.length > 0;
  };

  const handleSubmit = () => {
    if (!validateId(expertId)) {
      setError('ID must contain only lowercase letters, numbers, and hyphens');
      return;
    }

    createExpert(expertId);
    setExpertId('');
    setError('');
    onClose();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSubmit();
    } else if (e.key === 'Escape') {
      onClose();
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-vscode-editor-bg rounded-lg p-6 w-96 shadow-xl border border-vscode-border">
        <h2 className="text-lg font-bold mb-4 text-vscode-foreground">Create New Expert</h2>

        <div className="mb-4">
          <label className="block text-sm font-medium mb-1 text-vscode-foreground">
            Expert ID
          </label>
          <input
            type="text"
            value={expertId}
            onChange={(e) => {
              setExpertId(e.target.value.toLowerCase());
              setError('');
            }}
            onKeyDown={handleKeyDown}
            placeholder="e.g., backend-go-gin"
            className="w-full p-2 bg-vscode-input-bg text-vscode-input-fg border border-vscode-border rounded focus:outline-none focus:border-vscode-focus-border"
            autoFocus
          />
          {error && <p className="text-red-400 text-sm mt-1">{error}</p>}
          <p className="text-vscode-foreground opacity-60 text-xs mt-1">
            Lowercase letters, numbers, and hyphens only
          </p>
        </div>

        <div className="flex justify-end gap-2">
          <button
            onClick={onClose}
            className="px-4 py-2 text-vscode-foreground hover:bg-vscode-list-hover rounded"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            className="px-4 py-2 bg-vscode-button-bg text-vscode-button-fg rounded hover:bg-vscode-button-hover-bg"
          >
            Create
          </button>
        </div>
      </div>
    </div>
  );
};

export default CreateExpertDialog;
