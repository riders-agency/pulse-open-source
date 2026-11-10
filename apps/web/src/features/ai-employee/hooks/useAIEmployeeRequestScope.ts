import { useCallback, useLayoutEffect, useRef } from "react";

// Late responses must never replace another project's chat or account state.
export function useAIEmployeeRequestScope(projectId?: string | null, token?: string | null) {
  const epoch = useRef(0);
  useLayoutEffect(() => {
    epoch.current++;
    return () => { epoch.current++; };
  }, [projectId, token]);
  return useCallback(() => {
    const started = epoch.current;
    return () => started === epoch.current;
  }, []);
}
