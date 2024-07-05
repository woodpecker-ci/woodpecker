import Notifications, { notify, type NotificationsOptions } from '@kyvg/vue3-notification';

export const notifications = Notifications;

function notifyError(err: Error, args: NotificationsOptions | string = {}): void {
  console.error(err);

  const mArgs = typeof args === 'string' ? { title: args } : args;
  const title = mArgs?.title ?? err?.message ?? err?.toString();

  notify({ type: 'error', ...mArgs, title });
}

export default () => ({ notify, notifyError });
