import { useTranslation } from 'react-i18next';
import { Alert, Modal } from 'antd';

import { QrPanel } from '@/pages/inbounds/qr';
import { clashSubscriptionUrl } from './clashSubscription';

interface ClashQrModalProps {
  open: boolean;
  email: string;
  subId: string | undefined;
  subSettings: {
    enable: boolean;
    subClashEnable: boolean;
    subClashURI: string;
  };
  onOpenChange: (open: boolean) => void;
}

export default function ClashQrModal({
  open,
  email,
  subId,
  subSettings,
  onOpenChange,
}: ClashQrModalProps) {
  const { t } = useTranslation();
  const url = clashSubscriptionUrl(subId, subSettings.subClashURI);
  const unavailableReason = !subSettings.enable
    ? t('pages.clients.subLinksDisabled')
    : !subSettings.subClashEnable
      ? t('pages.clients.clashSubscriptionDisabled')
      : !subId
        ? t('pages.clients.clashSubscriptionMissingId')
        : !url
          ? t('pages.clients.clashSubscriptionInvalidUrl')
          : null;
  return (
    <Modal
      open={open}
      title={`${t('pages.clients.clashSubscription')} - ${email}`}
      footer={null}
      width="fit-content"
      onCancel={() => onOpenChange(false)}
    >
      {unavailableReason ? (
        <Alert type="info" title={unavailableReason} showIcon />
      ) : (
        url && <QrPanel value={url} remark={t('pages.clients.clashSubscription')} />
      )}
    </Modal>
  );
}
