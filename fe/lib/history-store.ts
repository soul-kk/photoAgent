import type { AnalysisHistoryItem } from './api';

let _selected: AnalysisHistoryItem | null = null;

export function setSelectedHistoryItem(item: AnalysisHistoryItem) {
  _selected = item;
}

export function getSelectedHistoryItem(): AnalysisHistoryItem | null {
  return _selected;
}
