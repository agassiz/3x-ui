import { describe, expect, it } from 'vitest';

import { RandomUtil } from '@/utils';

describe('RandomUtil.randomShortIds', () => {
  it('never produces a shortId shorter than 6 hex chars', () => {
    for (let run = 0; run < 50; run++) {
      const ids = RandomUtil.randomShortIds().split(',');
      expect(ids.length).toBeGreaterThan(0);
      for (const id of ids) {
        expect(id).toMatch(/^[0-9a-f]{6,16}$/);
        expect(id.length % 2).toBe(0);
      }
    }
  });
});
