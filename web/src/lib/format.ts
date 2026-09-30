export const initials = (s: string) =>
  s
    .split(" ")
    .map((x) => x[0])
    .slice(0, 2)
    .join("");
export const date = (s: string) =>
  new Date(s).toLocaleDateString("ru-RU", { day: "numeric", month: "short" });
export const time = (s: string) =>
  new Date(s).toLocaleTimeString("ru-RU", {
    hour: "2-digit",
    minute: "2-digit",
  });
export const size = (n: number) =>
  n < 1024
    ? `${n} Б`
    : n < 1024 * 1024
      ? `${Math.round(n / 1024)} КБ`
      : `${(n / 1024 / 1024).toFixed(1)} МБ`;
