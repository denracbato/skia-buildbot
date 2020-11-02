// Declaration for the window.sk global variable.
import { SkPerfConfig } from '../json';

declare global {
    interface Window {
      sk: {
        perf: SkPerfConfig;
      };
    }
  }
