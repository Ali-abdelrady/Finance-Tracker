export function formatCurrency(
  cents: number,
  currency = "EGP",
  locale = "en-US",
) {
  return new Intl.NumberFormat(locale, {
    style: "currency",
    currency,
    maximumFractionDigits: 2,
  }).format((cents || 0) / 100);
}
export function toMinorUnits(value: string | number): number {
  const amount = typeof value === "number" ? value : Number(value);
  if (!Number.isFinite(amount)) throw new Error("Enter a valid amount.");
  return Math.round((amount + Number.EPSILON) * 100);
}
