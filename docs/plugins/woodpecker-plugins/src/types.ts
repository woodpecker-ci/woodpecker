export type WoodpeckerPluginHeader = {
  name?: string;
  description?: string;
  icon?: string;
};

export type WoodpeckerPlugin = {
  name: string;
  repoName: string;
  description: string;
  url: string;
  icon: string;
  docs: string;
};

export type Content = {
  plugins: WoodpeckerPlugin[];
};
