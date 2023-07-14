import { RouteLocationRaw, useRouter } from 'vue-router';

export function useRouteBackOrDefault(to: RouteLocationRaw) {
  const router = useRouter();

  return async () => {
    if ((window.history.state as { back: string }).back === null) {
      await router.replace(to);
      return;
    }
    router.back();
  };
}
