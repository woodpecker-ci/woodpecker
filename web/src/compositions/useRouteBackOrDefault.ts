import { RouteLocationRaw, useRouter } from 'vue-router';

export function useRouteBackOrDefault(to: RouteLocationRaw) {
  const router = useRouter();

  return () => {
    // TODO: use history navigation once we have found a solution for filtering external history entries
    // if (window.history.length > 2) {
    //   router.back();
    // } else {
    router.replace(to);
    // }
  };
}
