import { RouteLocationRaw, useRouter } from 'vue-router';

export function useRouteBackOrDefault(to: RouteLocationRaw, forceTo: boolean) {
  const router = useRouter();

  return async () => {
    if (forceTo === true) {
      await router.replace(to);
      return;
    }
    if ((window.history.state as { back: string }).back === null) {
      await router.replace(to);
      return;
    }
    router.back();
  };
}
