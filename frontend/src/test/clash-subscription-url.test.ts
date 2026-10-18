import { describe, expect, it } from 'vitest';

import { clashSubscriptionUrl } from '@/pages/clients/clashSubscription';

describe('clashSubscriptionUrl', () => {
  it.each([
    ['https://sub.example.com/clash/', 'https://sub.example.com/clash/token%2Fa%3Fb'],
    ['https://sub.example.com:2096/custom', 'https://sub.example.com:2096/custom/token%2Fa%3Fb'],
    ['http://[::1]:2096/custom/', 'http://[::1]:2096/custom/token%2Fa%3Fb'],
    ['https://sub.example.com/', 'https://sub.example.com/token%2Fa%3Fb'],
  ])('uses the configured subscription prefix %j', (baseURI, want) => {
    expect(clashSubscriptionUrl('token/a?b', baseURI)).toBe(want);
  });

  it.each(['', '/clash/', 'invalid', 'ftp://sub.example.com/clash/'])(
    'does not guess an address for an invalid prefix %j',
    (baseURI) => {
      expect(clashSubscriptionUrl('token', baseURI)).toBeNull();
    },
  );

  it.each(['', undefined])('does not generate a link without a subId %j', (subId) => {
    expect(clashSubscriptionUrl(subId, 'https://sub.example.com/clash/')).toBeNull();
  });
});
