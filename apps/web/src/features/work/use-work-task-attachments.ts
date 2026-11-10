import { useState } from "react";

import type { WorkTaskAttachment, WorkTaskDetail } from "../../lib/api";
import { fetchWorkTaskAttachment, uploadWorkTaskAttachment } from "./api/workApi";
import type { TaskTab } from "./work-task-types";

export function useWorkTaskAttachments({
  onChanged,
  setError,
  setTab,
  setTask,
  task,
  token,
}: {
  onChanged?: () => void;
  setError: (value: string) => void;
  setTab: (value: TaskTab) => void;
  setTask: (value: WorkTaskDetail | undefined) => void;
  task?: WorkTaskDetail;
  token?: string | null;
}) {
  const [uploading, setUploading] = useState(false);

  const uploadAttachment = async (file?: File) => {
    if (!token || !task || !file) return;
    setUploading(true);
    setError("");
    try {
      const form = new FormData();
      form.append("file", file);
      const uploaded = await uploadWorkTaskAttachment(task.task_id, token, form);
      setTask({
        ...task,
        attachments: [uploaded, ...(task.attachments || [])],
        attachments_count: (task.attachments_count || 0) + 1,
      });
      setTab("attachments");
      onChanged?.();
    } finally {
      setUploading(false);
    }
  };

  const downloadAttachment = async (attachment: WorkTaskAttachment) => {
    if (!token) return;
    const blob = await fetchWorkTaskAttachment(attachment, token);
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = attachment.file_name;
    document.body.appendChild(a);
    a.click();
    a.remove();
    window.URL.revokeObjectURL(url);
  };

  return { downloadAttachment, uploadAttachment, uploading };
}
