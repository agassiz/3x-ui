import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import ClashQrModal from '@/pages/clients/ClashQrModal';
import { renderWithProviders } from './test-utils';

vi.mock('@/pages/inbounds/qr', () => ({
  QrPanel: ({ value }: { value: string }) => <div data-testid="clash-qr-value">{value}</div>,
}));

const settings = {
  enable: true,
  subClashEnable: true,
  subClashURI: 'https://sub.example.com:2096/custom-clash/',
};

function subject(subSettings = settings, subId: string | undefined = 'private-token') {
  return (
    <ClashQrModal
      open
      email="alice@example.com"
      subId={subId}
      subSettings={subSettings}
      onOpenChange={vi.fn()}
    />
  );
}

describe('ClashQrModal', () => {
  it('shows the subscription service URL with the subId instead of the email', () => {
    renderWithProviders(subject());
    const value = screen.getByTestId('clash-qr-value').textContent;
    expect(value).toBe('https://sub.example.com:2096/custom-clash/private-token');
    expect(value).not.toContain('alice');
    expect(value).not.toContain('clash/subscription');
  });

  it.each([
    [{ ...settings, enable: false }, 'private-token', 'Subscription service is disabled.'],
    [
      { ...settings, subClashEnable: false },
      'private-token',
      'Clash subscriptions are disabled. Enable Clash in Panel Settings → Subscription.',
    ],
    [settings, '', 'This client has no subscription ID.'],
    [
      { ...settings, subClashURI: '' },
      'private-token',
      'The Clash subscription URL is missing or invalid. Check Panel Settings → Subscription.',
    ],
    [
      { ...settings, subClashURI: '/clash/' },
      'private-token',
      'The Clash subscription URL is missing or invalid. Check Panel Settings → Subscription.',
    ],
  ])(
    'does not offer a QR code when subscription configuration is unavailable: %j',
    (config, subId, reason) => {
      renderWithProviders(subject(config, subId));
      expect(screen.queryByTestId('clash-qr-value')).toBeNull();
      expect(screen.getByText(reason)).toBeTruthy();
    },
  );

  it('updates the open QR code when the token changes and removes it when disabled', () => {
    const view = renderWithProviders(subject());
    view.rerender(subject(settings, 'new-token'));
    expect(screen.getByTestId('clash-qr-value').textContent).toBe(
      'https://sub.example.com:2096/custom-clash/new-token',
    );
    view.rerender(subject({ ...settings, subClashEnable: false }));
    expect(screen.queryByTestId('clash-qr-value')).toBeNull();
  });
});
