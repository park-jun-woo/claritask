import { useParams, useNavigate } from 'react-router-dom'

/**
 * URL params.projectId와 서버 상태를 동기화하는 훅.
 * - ProjectLayout 내부: URL에서 projectId 추출
 * - 외부: GLOBAL 모드
 */
export function useProjectContext() {
  const { projectId } = useParams<{ projectId?: string }>()
  const navigate = useNavigate()

  const isGlobal = !projectId

  const navigateToProject = (id: string) => {
    if (id === 'none' || id === 'GLOBAL') {
      navigate('/')
    } else {
      navigate(`/projects/${id}/tasks`)
    }
  }

  return {
    projectId: projectId || 'GLOBAL',
    isGlobal,
    navigateToProject,
  }
}
