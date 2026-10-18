import { describe, expect, it } from 'vitest';
import { Form } from 'antd';
import { fireEvent, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { FormProvider, useForm, useWatch } from 'react-hook-form';

import { RealityForm } from '@/pages/inbounds/form/security';
import { RealityClientSettingsSchema } from '@/schemas/protocols/security/reality';
import type { InboundFormValues } from '@/schemas/forms/inbound-form';
import { renderWithProviders } from './test-utils';

describe('Reality uTLS custom value', () => {
  it('keeps a client-only fingerprint such as chrome140 instead of rejecting it', () => {
    const parsed = RealityClientSettingsSchema.parse({ fingerprint: 'chrome140' });
    expect(parsed.fingerprint).toBe('chrome140');
  });

  it('still defaults an unset fingerprint to chrome', () => {
    expect(RealityClientSettingsSchema.parse({}).fingerprint).toBe('chrome');
  });

  it('lets the form field take a typed custom value', () => {
    function FingerprintProbe() {
      const value = useWatch<InboundFormValues>({
        name: 'streamSettings.realitySettings.settings.fingerprint',
      });
      return <output data-testid="fingerprint">{String(value)}</output>;
    }
    function Harness({ children }: { children: ReactNode }) {
      const methods = useForm<InboundFormValues>({
        defaultValues: {
          streamSettings: { realitySettings: { settings: { fingerprint: 'chrome' } } },
        } as never,
      });
      return (
        <FormProvider {...methods}>
          <Form>{children}</Form>
        </FormProvider>
      );
    }
    const noop = () => {};
    renderWithProviders(
      <Harness>
        <FingerprintProbe />
        <RealityForm
          saving={false}
          scanning={false}
          scanResult={null}
          scanRealityTarget={noop}
          scanRealityCandidates={async () => []}
          applyRealityScanResult={noop}
          randomizeShortIds={noop}
          randomizeSpiderX={noop}
          genRealityKeypair={noop}
          clearRealityKeypair={noop}
          genMldsa65={noop}
          clearMldsa65={noop}
        />
      </Harness>,
    );
    const input = screen
      .getAllByRole('combobox')
      .find((el) => el.getAttribute('value') === 'chrome');
    expect(input).toBeDefined();
    fireEvent.change(input as HTMLElement, { target: { value: 'chrome140' } });
    expect(screen.getByTestId('fingerprint').textContent).toBe('chrome140');
  });
});
