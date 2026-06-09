const rawBuildTimestamp = import.meta.env.VITE_BUILD_TIMESTAMP || '';

export const buildInfo = {
  version: import.meta.env.VITE_APP_VERSION || 'v1.3',
  buildTimestamp: rawBuildTimestamp,
  buildLabel: rawBuildTimestamp ? rawBuildTimestamp.replace('T', ' ').replace(/\.\d{3}Z$/, ' UTC') : 'development',
};
