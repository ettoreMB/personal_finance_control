const MAX_DIGITS = 12;

function formatThousands(value: number): string {
  return String(value).replace(/\B(?=(\d{3})+(?!\d))/g, ".");
}

export function formatReaisFromCents(cents: number): string {
  const value = Number.isFinite(cents) ? Math.trunc(Math.abs(cents)) : 0;
  const whole = Math.floor(value / 100);
  const fraction = value % 100;
  return `${formatThousands(whole)},${String(fraction).padStart(2, "0")}`;
}

export function maskReaisInput(raw: string): string {
  const digits = raw.replace(/\D/g, "").slice(0, MAX_DIGITS);
  if (digits.length === 0) {
    return "";
  }
  return formatReaisFromCents(Number(digits));
}

export function parseReaisToCents(raw: string): number {
  const digits = raw.replace(/\D/g, "");
  if (digits.length === 0) {
    return 0;
  }
  return Number(digits);
}
