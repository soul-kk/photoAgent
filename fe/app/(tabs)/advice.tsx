import React from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  TextInput,
  Image,
  ActivityIndicator,
  StyleSheet,
  SafeAreaView,
  Alert,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import * as ImagePicker from 'expo-image-picker';
import { useRouter } from 'expo-router';
import { useAuth } from '@/context/auth';
import { TabBar } from '@/components/tab-bar';

// ─── Types ────────────────────────────────────────────────────────────────────

type Step = 1 | 2 | 3;

interface AdviceResult {
  position: { angle: string; description: string };
  focal: { value: string; description: string };
  tips: string[];
  alt: string;
}

// ─── Mock API (replace with real call later) ──────────────────────────────────

async function fetchAdvice(
  _imageUri: string,
  _subject: string,
  _token: string | null,
): Promise<AdviceResult> {
  // TODO: replace with actual API call
  // e.g. const formData = new FormData(); formData.append('image', ...); formData.append('subject', subject);
  // return await fetch('https://your-api/advice', { method: 'POST', body: formData }).then(r => r.json());

  // Simulate network delay then return mock data
  await new Promise((r) => setTimeout(r, 99999999)); // stays loading until real API is ready
  return {} as AdviceResult;
}

// ─── Step indicator ───────────────────────────────────────────────────────────

const STEPS = [
  { n: '1', label: '上传场景' },
  { n: '2', label: '描述主体' },
  { n: '3', label: '获取建议' },
];

function StepIndicator({ current }: { current: Step }) {
  return (
    <View style={styles.steps}>
      {STEPS.map((s, i) => {
        const stepNum = (i + 1) as Step;
        const done = stepNum < current;
        const active = stepNum === current;
        return (
          <React.Fragment key={s.n}>
            <View style={styles.step}>
              <View
                style={[
                  styles.stepDot,
                  done && styles.stepDotDone,
                  !active && !done && styles.stepDotInactive,
                ]}
              >
                {done ? (
                  <Ionicons name="checkmark" size={12} color="#0A0A0A" />
                ) : (
                  <Text style={[styles.stepNum, !active && styles.stepNumInactive]}>
                    {s.n}
                  </Text>
                )}
              </View>
              <Text style={[styles.stepLabel, !active && !done && styles.stepLabelInactive]}>
                {s.label}
              </Text>
            </View>
            {i < 2 && (
              <View style={[styles.stepLine, done && styles.stepLineDone]} />
            )}
          </React.Fragment>
        );
      })}
    </View>
  );
}

// ─── Styles ───────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#0A0A0A' },
  header: {
    height: 56,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 24,
    borderBottomWidth: 1,
    borderBottomColor: '#2A2A2A',
  },
  headerTitle: { fontFamily: 'PlayfairDisplay_700Bold', fontSize: 18, color: '#FFFFFF' },
  headerSub: { fontFamily: 'DMMono_400Regular', fontSize: 11, color: '#555555' },
  scroll: { flex: 1 },
  scrollContent: { padding: 24, paddingBottom: 120, gap: 24 },

  // Steps
  steps: { flexDirection: 'row', alignItems: 'center' },
  step: { alignItems: 'center', gap: 4 },
  stepDot: {
    width: 24, height: 24, borderRadius: 12, backgroundColor: '#FFFFFF',
    alignItems: 'center', justifyContent: 'center',
  },
  stepDotDone: { backgroundColor: '#FFFFFF' },
  stepDotInactive: { backgroundColor: '#1E1E1E', borderWidth: 1, borderColor: '#333333' },
  stepNum: { fontFamily: 'DMMono_500Medium', fontSize: 10, fontWeight: '700', color: '#0A0A0A' },
  stepNumInactive: { color: '#555555' },
  stepLabel: { fontFamily: 'DMMono_400Regular', fontSize: 9, color: '#FFFFFF', letterSpacing: 1 },
  stepLabelInactive: { color: '#333333' },
  stepLine: { flex: 1, height: 1, backgroundColor: '#2A2A2A', marginBottom: 16 },
  stepLineDone: { backgroundColor: '#FFFFFF' },

  // Scene photo
  scenePhoto: {
    height: 180, backgroundColor: '#141414', borderRadius: 4,
    borderWidth: 1, borderColor: '#2A2A2A',
    alignItems: 'center', justifyContent: 'center', gap: 8, overflow: 'hidden',
  },
  scenePhotoFilled: { borderColor: '#333333' },
  sceneImage: { position: 'absolute', top: 0, left: 0, right: 0, bottom: 0 },
  sceneReplace: {
    position: 'absolute', bottom: 12, right: 12,
    flexDirection: 'row', alignItems: 'center', gap: 4,
    backgroundColor: '#00000080', borderRadius: 2, paddingHorizontal: 10, paddingVertical: 5,
  },
  sceneReplaceText: { fontFamily: 'DMMono_400Regular', fontSize: 10, color: '#FFFFFF' },
  scenePromptTitle: { fontFamily: 'DMSans_500Medium', fontSize: 14, color: '#555555' },
  scenePromptSub: { fontFamily: 'DMMono_400Regular', fontSize: 10, color: '#333333' },

  // Input card
  inputCard: {
    backgroundColor: '#141414', borderRadius: 4, borderWidth: 1,
    borderColor: '#2A2A2A', padding: 18, gap: 10,
  },
  inputCardDisabled: { opacity: 0.4 },
  inputLabel: { fontFamily: 'DMMono_400Regular', fontSize: 10, color: '#555555', letterSpacing: 2 },
  inputField: { fontFamily: 'DMSans_500Medium', fontSize: 15, color: '#FFFFFF', paddingVertical: 0 },
  inputUnderline: { height: 1, backgroundColor: '#2A2A2A' },
  inputUnderlineActive: { backgroundColor: '#FFFFFF' },

  // Submit
  submitBtn: {
    height: 52, backgroundColor: '#FFFFFF', borderRadius: 4,
    flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 8,
  },
  submitBtnDisabled: { opacity: 0.35 },
  submitBtnText: { fontFamily: 'DMSans_700Bold', fontSize: 15, color: '#0A0A0A' },

  divider: { height: 1, backgroundColor: '#2A2A2A' },
  sectionLabel: { fontFamily: 'DMMono_400Regular', fontSize: 10, color: '#555555', letterSpacing: 2 },

  // Cards
  card: {
    backgroundColor: '#141414', borderRadius: 4, borderWidth: 1,
    borderColor: '#2A2A2A', padding: 18, gap: 12,
  },
  cardTopRow: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  cardTitleRow: { flexDirection: 'row', alignItems: 'center', gap: 6 },
  cardTitle: { fontFamily: 'DMSans_700Bold', fontSize: 14, color: '#FFFFFF' },
  cardText: { fontFamily: 'DMSans_400Regular', fontSize: 13, color: '#999999', lineHeight: 22 },
  smallTag: {
    backgroundColor: '#FFFFFF1A', borderRadius: 2, paddingHorizontal: 8, paddingVertical: 3,
  },
  smallTagText: { fontFamily: 'DMMono_400Regular', fontSize: 10, color: '#999999' },
  focalValue: { fontFamily: 'PlayfairDisplay_700Bold', fontSize: 22, color: '#FFFFFF' },
  diagram: {
    height: 80, backgroundColor: '#1A1A1A', borderRadius: 4,
    borderWidth: 1, borderColor: '#2A2A2A', alignItems: 'center', justifyContent: 'center',
  },
  diagramText: { fontFamily: 'DMMono_400Regular', fontSize: 11, color: '#333333' },
  tipRow: { flexDirection: 'row', alignItems: 'flex-start', gap: 10 },
  tipDot: { width: 6, height: 6, borderRadius: 3, backgroundColor: '#FFFFFF', marginTop: 7 },
  tipText: { flex: 1, fontFamily: 'DMSans_400Regular', fontSize: 13, color: '#999999', lineHeight: 22 },
  altSub: { fontFamily: 'DMMono_400Regular', fontSize: 11, color: '#555555', letterSpacing: 1 },

  // Loading
  loadingRow: { flexDirection: 'row', alignItems: 'center', gap: 10 },
  loadingText: { fontFamily: 'DMMono_400Regular', fontSize: 12, color: '#555555' },
  loadingBar: { height: 8, backgroundColor: '#1E1E1E', borderRadius: 2, width: '100%' },

  // Reset
  resetBtn: {
    flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 6,
    paddingVertical: 14,
  },
  resetBtnText: { fontFamily: 'DMMono_400Regular', fontSize: 12, color: '#555555' },
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

// ─── Main Screen ──────────────────────────────────────────────────────────────

export default function AdviceScreen() {
  const { token } = useAuth();
  const router = useRouter();
  const [step, setStep] = React.useState<Step>(1);
  const [imageUri, setImageUri] = React.useState<string | null>(null);
  const [subject, setSubject] = React.useState('');
  const [loading, setLoading] = React.useState(false);
  const [result, setResult] = React.useState<AdviceResult | null>(null);

  // Step 1 — pick image
  async function handlePickImage() {
    const { status } = await ImagePicker.requestMediaLibraryPermissionsAsync();
    if (status !== 'granted') {
      Alert.alert('需要相册权限', '请在设置中允许访问相册');
      return;
    }
    const picked = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ImagePicker.MediaTypeOptions.Images,
      quality: 0.85,
    });
    if (!picked.canceled && picked.assets.length > 0) {
      setImageUri(picked.assets[0].uri);
      setStep(2);
    }
  }

  // Step 2 → 3 — submit
  async function handleSubmit() {
    if (!imageUri || !subject.trim()) return;
    setStep(3);
    setLoading(true);
    try {
      const data = await fetchAdvice(imageUri, subject.trim(), token);
      setResult(data);
    } catch {
      Alert.alert('请求失败', '请稍后重试');
      setStep(2);
    } finally {
      setLoading(false);
    }
  }

  function handleReset() {
    setStep(1);
    setImageUri(null);
    setSubject('');
    setResult(null);
    setLoading(false);
  }

  // Auth guard
  if (!token) {
    return (
      <SafeAreaView style={styles.root}>
        <View style={styles.header}>
          <Text style={styles.headerTitle}>AI 拍摄建议</Text>
          <Text style={styles.headerSub}>拍前规划</Text>
        </View>
        <View style={styles.authGate}>
          <Ionicons name="lock-closed-outline" size={32} color="#555555" />
          <Text style={styles.authGateTitle}>需要登录</Text>
          <Text style={styles.authGateText}>请登录后使用 AI 拍摄建议功能</Text>
          <TouchableOpacity style={styles.authGateBtn} onPress={() => router.push('/auth' as any)}>
            <Text style={styles.authGateBtnText}>去登录</Text>
          </TouchableOpacity>
        </View>
        <TabBar />
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.root}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.headerTitle}>AI 拍摄建议</Text>
        <Text style={styles.headerSub}>拍前规划</Text>
      </View>

      <ScrollView
        style={styles.scroll}
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
        keyboardShouldPersistTaps="handled"
      >
        {/* Steps */}
        <StepIndicator current={step} />

        {/* ── Step 1: Scene photo ── */}
        <TouchableOpacity
          style={[styles.scenePhoto, imageUri && styles.scenePhotoFilled]}
          onPress={handlePickImage}
          activeOpacity={0.8}
        >
          {imageUri ? (
            <>
              <Image source={{ uri: imageUri }} style={styles.sceneImage} />
              <TouchableOpacity style={styles.sceneReplace} onPress={handlePickImage}>
                <Ionicons name="repeat-outline" size={14} color="#FFFFFF" />
                <Text style={styles.sceneReplaceText}>重新上传</Text>
              </TouchableOpacity>
            </>
          ) : (
            <>
              <Ionicons name="image-outline" size={32} color="#555555" />
              <Text style={styles.scenePromptTitle}>点击上传场景照片</Text>
              <Text style={styles.scenePromptSub}>JPG · PNG · HEIF</Text>
            </>
          )}
        </TouchableOpacity>

        {/* ── Step 2: Subject input ── */}
        <View style={[styles.inputCard, step < 2 && styles.inputCardDisabled]}>
          <Text style={styles.inputLabel}>想拍摄的主体</Text>
          <TextInput
            style={styles.inputField}
            placeholder="例如：咖啡馆窗边的人像"
            placeholderTextColor="#444444"
            value={subject}
            onChangeText={setSubject}
            editable={step >= 2}
            returnKeyType="done"
            multiline={false}
          />
          <View style={[styles.inputUnderline, subject.length > 0 && styles.inputUnderlineActive]} />
        </View>

        {/* Submit button — visible on step 2 */}
        {step === 2 && (
          <TouchableOpacity
            style={[styles.submitBtn, !subject.trim() && styles.submitBtnDisabled]}
            onPress={handleSubmit}
            activeOpacity={0.85}
            disabled={!subject.trim()}
          >
            <Text style={styles.submitBtnText}>获取 AI 建议</Text>
            <Ionicons name="arrow-forward" size={16} color="#0A0A0A" />
          </TouchableOpacity>
        )}

        {/* ── Step 3: Results ── */}
        {step === 3 && (
          <>
            <View style={styles.divider} />
            <Text style={styles.sectionLabel}>AI 拍摄建议</Text>

            {loading ? (
              <LoadingCards />
            ) : result ? (
              <ResultCards result={result} onReset={handleReset} />
            ) : null}
          </>
        )}
      </ScrollView>

      <TabBar />
    </SafeAreaView>
  );
}

// ─── Loading skeleton ─────────────────────────────────────────────────────────

function LoadingCards() {
  return (
    <>
      {[1, 2, 3].map((i) => (
        <View key={i} style={styles.card}>
          <View style={styles.loadingRow}>
            <ActivityIndicator size="small" color="#555555" />
            <Text style={styles.loadingText}>AI 分析中…</Text>
          </View>
          <View style={styles.loadingBar} />
          <View style={[styles.loadingBar, { width: '60%' }]} />
        </View>
      ))}
    </>
  );
}

// ─── Result cards ─────────────────────────────────────────────────────────────

function ResultCards({ result, onReset }: { result: AdviceResult; onReset: () => void }) {
  return (
    <>
      {/* Position */}
      <View style={styles.card}>
        <View style={styles.cardTopRow}>
          <View style={styles.cardTitleRow}>
            <Ionicons name="location-outline" size={16} color="#FFFFFF" />
            <Text style={styles.cardTitle}>推荐机位</Text>
          </View>
          <View style={styles.smallTag}>
            <Text style={styles.smallTagText}>{result.position.angle}</Text>
          </View>
        </View>
        <Text style={styles.cardText}>{result.position.description}</Text>
        <View style={styles.diagram}>
          <Text style={styles.diagramText}>[ 机位示意图 ]</Text>
        </View>
      </View>

      {/* Focal */}
      <View style={styles.card}>
        <View style={styles.cardTopRow}>
          <View style={styles.cardTitleRow}>
            <Ionicons name="aperture-outline" size={16} color="#FFFFFF" />
            <Text style={styles.cardTitle}>建议焦段</Text>
          </View>
          <Text style={styles.focalValue}>{result.focal.value}</Text>
        </View>
        <Text style={styles.cardText}>{result.focal.description}</Text>
      </View>

      {/* Tips */}
      <View style={styles.card}>
        <Text style={styles.cardTitle}>拍摄要点</Text>
        {result.tips.map((tip, i) => (
          <View key={i} style={styles.tipRow}>
            <View style={styles.tipDot} />
            <Text style={styles.tipText}>{tip}</Text>
          </View>
        ))}
      </View>

      {/* Alt */}
      <View style={styles.card}>
        <View style={styles.cardTopRow}>
          <Text style={styles.cardTitle}>备选方案</Text>
        </View>
        <Text style={styles.altSub}>{result.alt}</Text>
      </View>

      {/* Reset */}
      <TouchableOpacity style={styles.resetBtn} onPress={onReset} activeOpacity={0.8}>
        <Ionicons name="refresh-outline" size={14} color="#555555" />
        <Text style={styles.resetBtnText}>重新拍摄</Text>
      </TouchableOpacity>
    </>
  );
}
