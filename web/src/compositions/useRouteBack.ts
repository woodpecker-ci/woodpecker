import { RouteLocationRaw, useRouter } from 'vue-router';

export function useRouteBack(to: RouteLocationRaw) {
  const router = useRouter();

  return async () => {
    await router.replace(to);
  };
}
