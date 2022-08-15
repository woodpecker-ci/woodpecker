import Notifications, { NotificationsOptions, notify } from '@kyvg/vue3-notification';

export const notifications = Notifications;

function notifyError(err: unknown, args: NotificationsOptions | string = {}): void {
  // eslint-disable-next-line no-console
  console.error(err);

  const mArgs = typeof args === 'string' ? { title: args } : args;
  const title = mArgs?.title || (err as Error)?.message || `${err}`;

  notify({ type: 'error', ...mArgs, title });
}

export default () => ({ notify, notifyError });
