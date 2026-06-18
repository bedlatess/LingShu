export const perKToM = (perK: string | number) => (Number(perK || 0) * 1000).toFixed(6);

export const perMToK = (perM: string | number) => (Number(perM || 0) / 1000).toFixed(8);
