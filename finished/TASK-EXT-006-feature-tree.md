# TASK-EXT-006: Feature Tree 컴포넌트

## 목표
Feature/Task 계층 구조를 보여주는 트리 뷰 컴포넌트 구현.

## 파일
`webview-ui/src/components/FeatureTree.tsx`

## 구현 내용

```typescript
import React, { useState } from 'react';
import { Feature, Task } from '../stores/projectStore';

interface FeatureTreeProps {
  features: Feature[];
  tasks: Task[];
}

export function FeatureTree({ features, tasks }: FeatureTreeProps) {
  const [expandedFeatures, setExpandedFeatures] = useState<Set<number>>(new Set());

  const toggleFeature = (featureId: number) => {
    setExpandedFeatures((prev) => {
      const next = new Set(prev);
      if (next.has(featureId)) {
        next.delete(featureId);
      } else {
        next.add(featureId);
      }
      return next;
    });
  };

  const getTasksByFeature = (featureId: number): Task[] => {
    return tasks.filter((t) => t.feature_id === featureId);
  };

  const getStatusIcon = (status: string): string => {
    switch (status) {
      case 'pending':
        return '○';
      case 'doing':
        return '◐';
      case 'done':
        return '●';
      case 'failed':
        return '✕';
      default:
        return '?';
    }
  };

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'pending':
        return 'text-gray-400';
      case 'doing':
        return 'text-blue-400';
      case 'done':
        return 'text-green-400';
      case 'failed':
        return 'text-red-400';
      default:
        return 'text-gray-400';
    }
  };

  return (
    <div className="p-2 text-sm">
      <div className="mb-2 px-2 text-xs text-gray-500 uppercase tracking-wide">
        Features
      </div>

      {features.length === 0 ? (
        <div className="px-2 text-gray-500 italic">No features</div>
      ) : (
        <ul className="space-y-1">
          {features.map((feature) => {
            const featureTasks = getTasksByFeature(feature.id);
            const isExpanded = expandedFeatures.has(feature.id);
            const doneCount = featureTasks.filter((t) => t.status === 'done').length;

            return (
              <li key={feature.id}>
                {/* Feature Row */}
                <div
                  className="flex items-center px-2 py-1 rounded cursor-pointer hover:bg-gray-700"
                  onClick={() => toggleFeature(feature.id)}
                >
                  <span className="mr-1 text-gray-500">
                    {isExpanded ? '▼' : '▶'}
                  </span>
                  <span className={`mr-2 ${getStatusColor(feature.status)}`}>
                    {getStatusIcon(feature.status)}
                  </span>
                  <span className="flex-1 truncate">{feature.name}</span>
                  <span className="text-xs text-gray-500">
                    {doneCount}/{featureTasks.length}
                  </span>
                </div>

                {/* Tasks */}
                {isExpanded && featureTasks.length > 0 && (
                  <ul className="ml-4 mt-1 space-y-1">
                    {featureTasks.map((task) => (
                      <li
                        key={task.id}
                        className="flex items-center px-2 py-1 rounded cursor-pointer hover:bg-gray-700"
                      >
                        <span className={`mr-2 ${getStatusColor(task.status)}`}>
                          {getStatusIcon(task.status)}
                        </span>
                        <span className="flex-1 truncate text-xs">
                          {task.title}
                        </span>
                      </li>
                    ))}
                  </ul>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
```

## 기능
- Feature 목록 표시
- 클릭으로 Task 목록 펼치기/접기
- 상태별 아이콘 및 색상
  - pending: 회색 ○
  - doing: 파란색 ◐
  - done: 녹색 ●
  - failed: 빨간색 ✕
- Feature별 완료 Task 카운트 표시

## 완료 조건
- [ ] FeatureTree 컴포넌트 구현
- [ ] Feature 펼치기/접기
- [ ] Task 상태별 색상
- [ ] 완료 카운트 표시
