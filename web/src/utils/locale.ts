export function getUserLanguage(): string {
  return navigator.language.split('-')[0];
}
