import React from 'react';
import {
  ActivityIndicator,
  Alert,
  Image,
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import * as ImagePicker from 'expo-image-picker';
import { useRouter } from 'expo-router';
import { useAuth } from '@/context/auth';
import { TabBar } from '@/components/tab-bar';
import {
  fetchPhotographyAnalyze,
  type AnalyzeFocusDimension,
  type PhotographyAnalyzeResponse,
} from '@/lib/api';

type DimensionKey = 'composition' | 'color' | 'exposure' | 'content';

type DimensionResult = {
  key: DimensionKey;
  label: string;
  icon: keyof typeof Ionicons.glyphMap;
  score: number;
  note: string;
};

type AnalysisResult = {
  overall_score: number;
  dimension_scores: Record<DimensionKey, number>;
  dimension_notes: Record<DimensionKey, string>;
  overall_analysis: string;
  improvement_tips: string[];
  focused_dimension?: string;
  focused_deep_analysis?: string;
};

const DIMENSION_META: Pick<DimensionResult, 'key' | 'label' | 'icon'>[] = [
  { key: 'composition', label: '构图', icon: 'grid-outline' },
  { key: 'color', label: '色彩', icon: 'color-palette-outline' },
  { key: 'exposure', label: '曝光', icon: 'sunny-outline' },
  { key: 'content', label: '内容识别', icon: 'scan-outline' },
];

function clampScore(value: unknown) {
  const n = typeof value === 'number' && Number.isFinite(value) ? Math.round(value) : 0;
  return Math.max(0, Math.min(100, n));
}

function getPickedFileName(asset: ImagePicker.ImagePickerAsset) {
  if (asset.fileName) return asset.fileName;
  const name = asset.uri.split('/').pop();
  return name && name.includes('.') ? name : 'analysis-photo.jpg';
}

function buildImagePart(asset: ImagePicker.ImagePickerAsset) {
  return {
    uri: asset.uri,
    name: getPickedFileName(asset),
    type: asset.mimeType ?? 'image/jpeg',
  };
}

function mapAnalyzeResponse(data: PhotographyAnalyzeResponse): AnalysisResult {
  return {
    overall_score: data.overall_score,
    dimension_scores: data.dimension_scores,
    dimension_notes: data.dimension_notes,
    overall_analysis: data.overall_analysis,
    improvement_tips: data.improvement_tips,
    focused_dimension: data.focused_dimension,
    focused_deep_analysis: data.focused_deep_analysis,
  };
}

function buildDimensions(data: AnalysisResult): DimensionResult[] {
  return DIMENSION_META.map((dim) => ({
    ...dim,
    score: clampScore(data.dimension_scores[dim.key]),
    note: data.dimension_notes[dim.key] ?? '暂无该维度评语，请稍后重试。',
  }));
}

async function requestAnalysis(
  asset: ImagePicker.ImagePickerAsset,
  token: string,
  focus?: AnalyzeFocusDimension,
) {
  const data = await fetchPhotographyAnalyze(token, asset.uri, {
    focusDimension: focus,
    mimeType: asset.mimeType ?? 'image/jpeg',
  });
  return mapAnalyzeResponse(data);
}

export default function AnalysisScreen() {
  const { token } = useAuth();
  const router = useRouter();
  const [photo, setPhoto] = React.useState<ImagePicker.ImagePickerAsset | null>(null);
  const [activeDim, setActiveDim] = React.useState<DimensionKey | null>(null);
  const [loading, setLoading] = React.useState(false);
  const [result, setResult] = React.useState<AnalysisResult | null>(null);
  const [error, setError] = React.useState<string | null>(null);

  const dimensions = React.useMemo(() => (result ? buildDimensions(result) : []), [result]);
  const visibleDimensions = activeDim
    ? dimensions.filter((dim) => dim.key === activeDim)
    : dimensions;

  // Auth guard
  if (!token) {
    return (
      <SafeAreaView style={styles.root}>
        <View style={styles.header}>
          <Text style={styles.headerTitle}>AI 摄影分析</Text>
          <Text style={styles.headerSub}>上传照片开始分析</Text>
        </View>
        <View style={styles.authGate}>
          <Ionicons name="lock-closed-outline" size={32} color="#555555" />
          <Text style={styles.authGateTitle}>需要登录</Text>
          <Text style={styles.authGateText}>请登录后使用 AI 分析功能</Text>
          <TouchableOpacity style={styles.authGateBtn} onPress={() => router.push('/auth' as any)}>
            <Text style={styles.authGateBtnText}>去登录</Text>
          </TouchableOpacity>
        </View>
        <TabBar />
      </SafeAreaView>
    );
  }

  async function handlePickImage() {
    const { status } = await ImagePicker.requestMediaLibraryPermissionsAsync();
    if (status !== 'granted') {
      Alert.alert('需要相册权限', '请在系统设置中允许访问照片后再上传。');
      return;
    }

    const picked = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ImagePicker.MediaTypeOptions.Images,
      quality: 0.9,
    });

    if (!picked.canceled && picked.assets.length > 0) {
      const asset = picked.assets[0];
      console.log('AI analysis selected file:', asset);
      setPhoto(asset);
      setResult(null);
      setActiveDim(null);
      setError(null);
    }
  }

  async function handleAnalyze(deepFocus = false) {
    if (!photo || loading || !token) return;

    setLoading(true);
    setError(null);
    if (!deepFocus) {
      setActiveDim(null);
    }
    try {
      const focus = deepFocus && activeDim ? activeDim : undefined;
      const data = await requestAnalysis(photo, token, focus);
      setResult(data);
      if (deepFocus && activeDim) {
        setActiveDim(activeDim);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : '分析失败，请稍后重试。';
      setError(message);
      Alert.alert('AI 分析失败', message);
    } finally {
      setLoading(false);
    }
  }

  function handleDimensionPress(key: DimensionKey) {
    setActiveDim((current) => (current === key ? null : key));
  }

  return (
    <SafeAreaView style={styles.root}>
      <View style={styles.header}>
        <Text style={styles.headerTitle}>AI 摄影分析</Text>
        <Text style={styles.headerSub}>上传照片开始分析</Text>
      </View>

      <ScrollView
        style={styles.scroll}
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
      >
        <TouchableOpacity
          style={[styles.dropzone, photo && styles.dropzoneFilled]}
          onPress={handlePickImage}
          activeOpacity={0.82}
        >
          {photo ? (
            <>
              <Image source={{ uri: photo.uri }} style={styles.photoPreview} />
              <View style={styles.replaceBadge}>
                <Ionicons name="repeat-outline" size={14} color="#FFFFFF" />
                <Text style={styles.replaceText}>重新上传</Text>
              </View>
            </>
          ) : (
            <>
              <Ionicons name="cloud-upload-outline" size={40} color="#555555" />
              <Text style={styles.dropzoneTitle}>上传照片</Text>
              <Text style={styles.dropzoneSub}>JPG · PNG · HEIF · 最大 20MB</Text>
            </>
          )}
        </TouchableOpacity>

        <TouchableOpacity
          style={[styles.analyzeBtn, (!photo || loading) && styles.analyzeBtnDisabled]}
          onPress={() => handleAnalyze(false)}
          activeOpacity={0.85}
          disabled={!photo || loading}
        >
          {loading ? (
            <ActivityIndicator size="small" color="#0A0A0A" />
          ) : (
            <Ionicons name="sparkles-outline" size={16} color="#0A0A0A" />
          )}
          <Text style={styles.analyzeBtnText}>{loading ? 'AI 分析中' : '开始 AI 分析'}</Text>
        </TouchableOpacity>

        <Text style={styles.sectionLabel}>分析维度</Text>
        <View style={styles.dimsRow}>
          {DIMENSION_META.map((dim) => {
            const active = activeDim === dim.key;
            return (
              <TouchableOpacity
                key={dim.key}
                style={[styles.dimTag, active && styles.dimTagActive]}
                onPress={() => handleDimensionPress(dim.key)}
                activeOpacity={0.8}
              >
                <Ionicons
                  name={dim.icon}
                  size={13}
                  color={active ? '#0A0A0A' : '#999999'}
                />
                <Text style={[styles.dimTagText, active && styles.dimTagTextActive]}>
                  {dim.label}
                </Text>
              </TouchableOpacity>
            );
          })}
        </View>

        {activeDim ? (
          <TouchableOpacity
            style={[styles.deepBtn, (!photo || loading) && styles.analyzeBtnDisabled]}
            onPress={() => handleAnalyze(true)}
            disabled={!photo || loading}
            activeOpacity={0.85}
          >
            <Text style={styles.deepBtnText}>
              深入分析：
              {DIMENSION_META.find((d) => d.key === activeDim)?.label}
            </Text>
          </TouchableOpacity>
        ) : null}

        <View style={styles.divider} />

        <Text style={styles.sectionLabel}>分析结果</Text>
        {loading ? (
          <LoadingState />
        ) : result ? (
          <>
            <View style={styles.scoresGrid}>
              {visibleDimensions.map((dim) => (
                <ScoreCard key={dim.key} item={dim} />
              ))}
            </View>

            <View style={styles.reportCard}>
              <Text style={styles.reportLabel}>
                {activeDim
                  ? `${visibleDimensions[0]?.label ?? ''}评语`
                  : `AI 分析报告 · 综合 ${clampScore(result.overall_score)} 分`}
              </Text>
              {visibleDimensions.map((dim) => (
                <View key={dim.key} style={styles.noteBlock}>
                  <View style={styles.noteTitleRow}>
                    <Text style={styles.noteTitle}>{dim.label}</Text>
                    <Text style={styles.noteScore}>{dim.score}</Text>
                  </View>
                  <Text style={styles.reportText}>{dim.note}</Text>
                </View>
              ))}
              {!activeDim ? (
                <>
                  <View style={styles.reportDivider} />
                  <Text style={styles.reportSuggest}>{result.overall_analysis}</Text>
                  {result.improvement_tips.length > 0 ? (
                    <>
                      <View style={styles.reportDivider} />
                      <Text style={styles.reportSuggest}>
                        {result.improvement_tips.map((t, i) => `${i + 1}. ${t}`).join('\n')}
                      </Text>
                    </>
                  ) : null}
                </>
              ) : null}
              {activeDim &&
              result.focused_dimension === activeDim &&
              result.focused_deep_analysis ? (
                <>
                  <View style={styles.reportDivider} />
                  <Text style={styles.reportLabel}>深入分析</Text>
                  <Text style={styles.reportSuggest}>{result.focused_deep_analysis}</Text>
                </>
              ) : null}
            </View>
          </>
        ) : (
          <View style={styles.emptyCard}>
            <Ionicons name="image-outline" size={22} color="#555555" />
            <Text style={styles.emptyText}>
              {error ?? '上传照片后点击开始分析，结果会显示在这里。'}
            </Text>
          </View>
        )}
      </ScrollView>

      <TabBar />
    </SafeAreaView>
  );
}

function LoadingState() {
  return (
    <View style={styles.loadingCard}>
      <ActivityIndicator size="small" color="#999999" />
      <View style={styles.loadingCopy}>
        <Text style={styles.loadingTitle}>正在读取画面细节</Text>
        <Text style={styles.loadingText}>构图、色彩、曝光和内容识别评分生成中</Text>
      </View>
    </View>
  );
}

function ScoreCard({ item }: { item: DimensionResult }) {
  return (
    <View style={styles.scoreCard}>
      <View style={styles.scoreTopRow}>
        <Text style={styles.scoreLabel}>{item.label}</Text>
        <Ionicons name={item.icon} size={14} color="#555555" />
      </View>
      <Text style={styles.scoreValue}>{item.score}</Text>
      <View style={styles.scoreBarBg}>
        <View style={[styles.scoreBarFill, { width: `${item.score}%` }]} />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  root: {
    flex: 1,
    backgroundColor: '#0A0A0A',
  },
  header: {
    height: 56,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 24,
    borderBottomWidth: 1,
    borderBottomColor: '#2A2A2A',
  },
  headerTitle: {
    fontFamily: 'PlayfairDisplay_700Bold',
    fontSize: 18,
    color: '#FFFFFF',
  },
  headerSub: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 11,
    color: '#555555',
  },
  scroll: {
    flex: 1,
  },
  scrollContent: {
    padding: 24,
    paddingBottom: 120,
    gap: 24,
  },
  dropzone: {
    height: 220,
    backgroundColor: '#111111',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#333333',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 10,
    overflow: 'hidden',
  },
  dropzoneFilled: {
    borderColor: '#555555',
  },
  photoPreview: {
    ...StyleSheet.absoluteFillObject,
  },
  replaceBadge: {
    position: 'absolute',
    right: 12,
    bottom: 12,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 5,
    backgroundColor: '#00000099',
    borderRadius: 2,
    paddingHorizontal: 10,
    paddingVertical: 6,
  },
  replaceText: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 10,
    color: '#FFFFFF',
  },
  dropzoneTitle: {
    fontFamily: 'PlayfairDisplay_700Bold',
    fontSize: 18,
    color: '#FFFFFF',
  },
  dropzoneSub: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 11,
    color: '#555555',
  },
  analyzeBtn: {
    height: 52,
    backgroundColor: '#FFFFFF',
    borderRadius: 4,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
  },
  analyzeBtnDisabled: {
    opacity: 0.35,
  },
  analyzeBtnText: {
    fontFamily: 'DMSans_700Bold',
    fontSize: 15,
    color: '#0A0A0A',
  },
  deepBtn: {
    borderWidth: 1,
    borderColor: '#444444',
    borderRadius: 4,
    paddingVertical: 12,
    alignItems: 'center',
  },
  deepBtnText: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 13,
    color: '#CCCCCC',
  },
  sectionLabel: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 10,
    color: '#555555',
    letterSpacing: 2,
  },
  dimsRow: {
    flexDirection: 'row',
    gap: 8,
    flexWrap: 'wrap',
  },
  dimTag: {
    backgroundColor: '#1E1E1E',
    borderWidth: 1,
    borderColor: '#333333',
    borderRadius: 2,
    paddingHorizontal: 12,
    paddingVertical: 7,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 5,
  },
  dimTagActive: {
    backgroundColor: '#FFFFFF',
    borderColor: '#FFFFFF',
  },
  dimTagText: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 12,
    color: '#999999',
  },
  dimTagTextActive: {
    fontFamily: 'DMSans_700Bold',
    color: '#0A0A0A',
  },
  divider: {
    height: 1,
    backgroundColor: '#2A2A2A',
  },
  scoresGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  scoreCard: {
    width: '48.8%',
    minHeight: 126,
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    padding: 14,
    paddingVertical: 16,
    gap: 8,
  },
  scoreTopRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  scoreLabel: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 10,
    color: '#555555',
    letterSpacing: 1,
  },
  scoreValue: {
    fontFamily: 'PlayfairDisplay_700Bold',
    fontSize: 34,
    color: '#FFFFFF',
  },
  scoreBarBg: {
    height: 2,
    backgroundColor: '#2A2A2A',
  },
  scoreBarFill: {
    height: 2,
    backgroundColor: '#FFFFFF',
  },
  reportCard: {
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    padding: 18,
    paddingVertical: 20,
    gap: 14,
  },
  reportLabel: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 10,
    color: '#555555',
    letterSpacing: 2,
  },
  noteBlock: {
    gap: 6,
  },
  noteTitleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  noteTitle: {
    fontFamily: 'DMSans_700Bold',
    fontSize: 13,
    color: '#FFFFFF',
  },
  noteScore: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 12,
    color: '#999999',
  },
  reportText: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 13,
    color: '#FFFFFF',
    lineHeight: 22,
  },
  reportDivider: {
    height: 1,
    backgroundColor: '#2A2A2A',
  },
  reportSuggest: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 13,
    color: '#999999',
    lineHeight: 22,
  },
  emptyCard: {
    minHeight: 120,
    backgroundColor: '#111111',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 10,
    padding: 20,
  },
  emptyText: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 13,
    color: '#555555',
    textAlign: 'center',
    lineHeight: 20,
  },
  loadingCard: {
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    padding: 18,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
  },
  loadingCopy: {
    flex: 1,
    gap: 4,
  },
  loadingTitle: {
    fontFamily: 'DMSans_700Bold',
    fontSize: 13,
    color: '#FFFFFF',
  },
  loadingText: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 12,
    color: '#555555',
  },
  authGate: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    gap: 12,
    paddingHorizontal: 40,
    paddingBottom: 80,
  },
  authGateTitle: {
    fontFamily: 'PlayfairDisplay_700Bold',
    fontSize: 20,
    color: '#FFFFFF',
  },
  authGateText: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 14,
    color: '#555555',
    textAlign: 'center',
  },
  authGateBtn: {
    marginTop: 8,
    height: 48,
    paddingHorizontal: 32,
    backgroundColor: '#FFFFFF',
    borderRadius: 4,
    alignItems: 'center',
    justifyContent: 'center',
  },
  authGateBtnText: {
    fontFamily: 'DMSans_700Bold',
    fontSize: 14,
    color: '#0A0A0A',
  },
});
