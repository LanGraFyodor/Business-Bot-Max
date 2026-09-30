import { FileText } from "lucide-react";
import type { Doc } from "../types";
const statuses: Record<string, { label: string; className: string }> = {
  draft: { label: "Черновик", className: "gray" },
  review: { label: "На согласовании", className: "amber" },
  approved: { label: "Согласован", className: "green" },
  changes: { label: "На доработке", className: "rose" },
};
export function Robot() {
  return (
    <span
      role="img"
      aria-label="Docspace assistant"
      className="robot robot-sprite"
    />
  );
}
export function Badge({ status }: { status: string }) {
  const s = statuses[status] || statuses.draft;
  return (
    <span className={`badge ${s.className}`}>
      <i />
      {s.label}
    </span>
  );
}
export function FileIcon({ doc }: { doc: Doc }) {
  const ext =
    doc.versions.at(-1)?.name.split(".").at(-1)?.toUpperCase() || "DOC";
  return (
    <span
      className={`file-icon ${ext === "PDF" ? "pdf" : ext === "XLSX" || ext === "CSV" ? "sheet" : ""}`}
    >
      <FileText size={25} />
      <small>{ext}</small>
    </span>
  );
}
