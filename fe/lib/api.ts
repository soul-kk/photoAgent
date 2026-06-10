import { API_BASE } from './config';

export type ApiEnvelope<T> = {
  code: number;
  err_code: number;
  message: string;
  data: T;
};

export type ShootAdviceAnnotation = {
  id: number;
  area: string;
  label: string;
  hint: string;
};

export type ShootAdvice = {
  scene_summary: string;
  scene_analysis: {
    light: string;
    space: string;
    background: string;
    atmosphere: string;
  };
  subject_plan: string;
  camera_position: {
    description: string;
    angle: string;
    distance: string;
    annotations: ShootAdviceAnnotation[];
  };
  focal_length: {
    range: string;
    category: string;
    reason: string;
  };
  shooting_tips: string[];
  alternatives: { style: string; description: string }[];
  summary: string;
};

export type ShootAdviceResponse = {
  rubric_id: string;
  model: string;
  subject: string;
  image?: Record<string, unknown>;
  advice: ShootAdvice;
};

async function parseJson<T>(res: Response): Promise<T> {
  const body = (await res.json()) as ApiEnvelope<T>;
  if (!res.ok || body.code !== 200) {
    throw new Error(body.message || `请求失败 ${res.status}`);
  }
  return body.data;
}

export async function login(account: string, password: string): Promise<string> {
  const res = await fetch(`${API_BASE}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ account, password }),
  });
  const data = await parseJson<{ access_token: string }>(res);
  return data.access_token;
}

export async function fetchShootAdvice(
  token: string,
  imageUri: string,
  subject: string,
  mimeType = 'image/jpeg',
): Promise<ShootAdviceResponse> {
  const form = new FormData();
  const name = imageUri.split('/').pop() || 'photo.jpg';
  form.append('image', {
    uri: imageUri,
    name,
    type: mimeType,
  } as unknown as Blob);
  form.append('subject', subject);

  const res = await fetch(`${API_BASE}/api/kimi/photography/shoot-advice`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: form,
  });
  return parseJson<ShootAdviceResponse>(res);
}

export type CompareDimensionScores = {
  composition: number;
  color: number;
  exposure: number;
  sharpness: number;
  creativity: number;
};

export type ComparePhotoResult = {
  index: number;
  overall_score: number;
  dimension_scores: CompareDimensionScores;
  pros: string;
  cons: string;
  image?: Record<string, unknown>;
};

export type CompareImagesResponse = {
  rubric_id: string;
  model: string;
  image_count: number;
  best_index: number;
  best_reason: string;
  summary: string;
  ranking: number[];
  photos: ComparePhotoResult[];
};

export async function fetchCompareImages(
  token: string,
  imageUris: string[],
): Promise<CompareImagesResponse> {
  const form = new FormData();
  imageUris.forEach((uri, i) => {
    const name = uri.split('/').pop() || `photo-${i + 1}.jpg`;
    form.append('images', {
      uri,
      name,
      type: 'image/jpeg',
    } as unknown as Blob);
  });

  const res = await fetch(`${API_BASE}/api/kimi/photography/compare-images`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: form,
  });
  return parseJson<CompareImagesResponse>(res);
}

export type ToneStyleParameters = {
  exposure: string;
  contrast: string;
  highlights: string;
  shadows: string;
  whites: string;
  blacks: string;
  saturation: string;
  vibrance: string;
  temperature: string;
  tint: string;
  hue_adjustments?: string;
  curve?: string;
  grain?: string;
};

export type ToneStylePayload = {
  style_name: string;
  style_match_summary: string;
  parameters: ToneStyleParameters;
  adjustment_notes: string[];
  before_after_description: string;
  preview_hints: string;
};

export type ToneStyleResponse = {
  rubric_id: string;
  model: string;
  style_description: string;
  has_reference_image: boolean;
  style: ToneStylePayload;
  image?: Record<string, unknown>;
};

export type AnalyzeDimensionScores = {
  composition: number;
  color: number;
  exposure: number;
  content: number;
};

export type AnalyzeDimensionNotes = {
  composition: string;
  color: string;
  exposure: string;
  content: string;
};

export type PhotographyAnalyzeResponse = {
  rubric_id: string;
  model: string;
  format: string;
  overall_score: number;
  dimension_scores: AnalyzeDimensionScores;
  dimension_notes: AnalyzeDimensionNotes;
  dimension_labels: Record<string, string>;
  overall_analysis: string;
  improvement_tips: string[];
  focused_dimension?: string;
  focused_deep_analysis?: string;
  focus_dimension?: string;
  focus_dimension_label?: string;
  image?: Record<string, unknown>;
};

export type AnalyzeFocusDimension =
  | 'composition'
  | 'color'
  | 'exposure'
  | 'content';

const FOCUS_DIM_MAP: Record<string, AnalyzeFocusDimension> = {
  构图: 'composition',
  色彩: 'color',
  曝光: 'exposure',
  内容识别: 'content',
};

export function focusDimensionFromLabel(label: string): AnalyzeFocusDimension | undefined {
  return FOCUS_DIM_MAP[label];
}

export async function fetchPhotographyAnalyze(
  token: string,
  imageUri: string,
  options?: {
    focusDimension?: AnalyzeFocusDimension;
    prompt?: string;
    mimeType?: string;
  },
): Promise<PhotographyAnalyzeResponse> {
  const form = new FormData();
  const name = imageUri.split('/').pop() || 'photo.jpg';
  form.append('image', {
    uri: imageUri,
    name,
    type: options?.mimeType ?? 'image/jpeg',
  } as unknown as Blob);
  if (options?.prompt?.trim()) {
    form.append('prompt', options.prompt.trim());
  }
  if (options?.focusDimension) {
    form.append('focus_dimension', options.focusDimension);
  }

  const res = await fetch(
    `${API_BASE}/api/kimi/photography/analyze-image?stream=false`,
    {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
      body: form,
    },
  );
  return parseJson<PhotographyAnalyzeResponse>(res);
}

export async function fetchToneStyle(
  token: string,
  styleDescription: string,
  imageUri?: string | null,
): Promise<ToneStyleResponse> {
  const form = new FormData();
  form.append('style_description', styleDescription);
  if (imageUri) {
    const name = imageUri.split('/').pop() || 'ref.jpg';
    form.append('image', {
      uri: imageUri,
      name,
      type: 'image/jpeg',
    } as unknown as Blob);
  }

  const res = await fetch(`${API_BASE}/api/kimi/photography/tone-style`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: form,
  });
  return parseJson<ToneStyleResponse>(res);
}
