import { describe, expect, it } from "vitest";

import {
  formatReaisFromCents,
  maskReaisInput,
  parseReaisToCents,
} from "./money";

describe("maskReaisInput", () => {
  it("formats digits as reais with two decimal places", () => {
    expect(maskReaisInput("1")).toBe("0,01");
    expect(maskReaisInput("11")).toBe("0,11");
    expect(maskReaisInput("110")).toBe("1,10");
    expect(maskReaisInput("11050")).toBe("110,50");
  });

  it("adds thousand separators", () => {
    expect(maskReaisInput("1234567")).toBe("12.345,67");
  });

  it("keeps only digits from a pasted formatted value", () => {
    expect(maskReaisInput("110,50")).toBe("110,50");
    expect(maskReaisInput("1.110,50")).toBe("1.110,50");
  });

  it("clears non-numeric input", () => {
    expect(maskReaisInput("")).toBe("");
    expect(maskReaisInput("abc")).toBe("");
  });

  it("caps the number of digits", () => {
    expect(maskReaisInput("123456789012345")).toBe("1.234.567.890,12");
  });
});

describe("formatReaisFromCents", () => {
  it("formats cents for editing", () => {
    expect(formatReaisFromCents(0)).toBe("0,00");
    expect(formatReaisFromCents(4500)).toBe("45,00");
    expect(formatReaisFromCents(11050)).toBe("110,50");
  });

  it("treats invalid numbers as zero", () => {
    expect(formatReaisFromCents(Number.NaN)).toBe("0,00");
    expect(formatReaisFromCents(Number.POSITIVE_INFINITY)).toBe("0,00");
    expect(formatReaisFromCents(-9900)).toBe("99,00");
  });
});

describe("parseReaisToCents", () => {
  it("parses masked reais into cents", () => {
    expect(parseReaisToCents("110,50")).toBe(11050);
    expect(parseReaisToCents("12.345,67")).toBe(1234567);
    expect(parseReaisToCents("0,00")).toBe(0);
  });

  it("returns 0 for empty input", () => {
    expect(parseReaisToCents("")).toBe(0);
    expect(parseReaisToCents("abc")).toBe(0);
  });
});
