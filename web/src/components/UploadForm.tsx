import {useState,useRef} from "react";
import {UploadCloud,Loader2} from "lucide-react";
import type {Channel} from "../types";
import {size} from "../lib/format";
export default function UploadForm({
  channels,
  initialChannel,
  busy,
  onSubmit,
  submitLabel = "Загрузить документ",
}: {
  submitLabel?: string;
  channels: Channel[];
  initialChannel: string;
  busy: boolean;
  onSubmit: (f: FormData) => void;
}) {
  const [file, setFile] = useState<File | null>(null),
    [title, setTitle] = useState(""),
    [drag, setDrag] = useState(false),
    [error, setError] = useState("");
  const input = useRef<HTMLInputElement>(null);
  function choose(f?: File) {
    if (!f) return;
    if (
      !/\.(pdf|docx|xlsx|txt|csv|png|jpe?g)$/i.test(f.name) ||
      f.size > 25 * 1024 * 1024 ||
      f.size === 0
    ) {
      setError("Выберите поддерживаемый файл размером до 25 МБ");
      return;
    }
    setError("");
    setFile(f);
    if (!title) setTitle(f.name.replace(/\.[^.]+$/, ""));
  }
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!file) {
          setError("Выберите файл");
          return;
        }
        const form = new FormData(e.currentTarget);
        form.set("file", file);
        onSubmit(form);
      }}
    >
      <p className="form-intro">Файл будет доступен участникам выбранной команды.</p>
      <button
        type="button"
        className={`dropzone ${drag ? "dragging" : ""}`}
        onClick={() => input.current?.click()}
        onDragOver={(e) => {
          e.preventDefault();
          setDrag(true);
        }}
        onDragLeave={() => setDrag(false)}
        onDrop={(e) => {
          e.preventDefault();
          setDrag(false);
          choose(e.dataTransfer.files[0]);
        }}
      >
        <UploadCloud size={31} />
        <strong>{file ? file.name : "Перетащите файл или выберите его"}</strong>
        <small>
          {file
            ? size(file.size)
            : "PDF, DOCX, XLSX, TXT, CSV, PNG, JPG · до 25 МБ"}
        </small>
      </button>
      <input
        ref={input}
        type="file"
        hidden
        accept=".pdf,.docx,.xlsx,.txt,.csv,.png,.jpg,.jpeg"
        onChange={(e) => choose(e.target.files?.[0])}
      />
      {error && (
        <p className="form-error" role="alert">
          {error}
        </p>
      )}
      <label className="field">
        Название документа
        <input
          name="title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
          maxLength={180}
          placeholder="Например, Договор с поставщиком"
        />
      </label>
      <div className="form-columns">
        <label className="field">
          Канал
          <select name="channelId" defaultValue={initialChannel}>
            {channels.map((c) => (
              <option key={c.id} value={c.id}>
                # {c.name}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          Срок согласования
          <input name="due" type="date" />
        </label>
      </div>
      <label className="field">
        Что нужно проверить?
        <textarea
          name="description"
          maxLength={2000}
          placeholder="Добавьте контекст для коллег (необязательно)"
        />
      </label>
      <button className="primary full" disabled={busy}>
        {busy ? (
          <Loader2 size={18} className="spin" />
        ) : (
          <UploadCloud size={18} />
        )}
        {submitLabel}
      </button>
    </form>
  );
}
