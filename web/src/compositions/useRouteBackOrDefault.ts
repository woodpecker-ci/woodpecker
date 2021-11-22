import { RouteLocationRaw, useRouter } from 'vue-router';

export function useRouteBackOrDefault(to: RouteLocationRaw) {
  const router = useRouter();

  return () => {
    if (window.history.length > 2) {
      router.back();
    } else {
      router.replace(to);
    }
  };
}
