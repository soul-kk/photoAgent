import * as ImagePicker from 'expo-image-picker';
import { Image } from 'expo-image';
import { useState } from 'react';
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  TextInput,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';
import { TabBar } from '@/components/tab-bar';
import {
  fetchCompareImages,
  login,
  type CompareImagesResponse,
} from '@/lib/api';

export default function CompareScreen() {
  const [token, setToken] = useState('');
  const [account, setAccount] = useState('');
  const [password, setPassword] = useState('');
  const [uris, setUris] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<CompareImagesResponse | null>(null);

  const pickImages = async () => {
    const perm = await ImagePicker.requestMediaLibraryPermissionsAsync();
    if (!perm.granted) {
      setError('需要相册权限');
      return;
    }
    const picked = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ['images'],
      allowsMultipleSelection: true,
      selectionLimit: 8,
      quality: 0.85,
    });
    if (!picked.canceled && picked.assets.length >= 2) {
      setUris(picked.assets.map((a) => a.uri));
      setError(null);
    } else if (!picked.canceled) {
      setError('请至少选择 2 张照片');
    }
  };

  const handleLogin = async () => {
    setLoading(true);
    setError(null);
    try {
      setToken(await login(account.trim(), password));
    } catch (e) {
      setError(e instanceof Error ? e.message : '登录失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCompare = async () => {
    if (!token) {
      setError('请先登录');
      return;
    }
    if (uris.length < 2) {
      setError('请至少选择 2 张照片');
      return;
    }
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      const res = await fetchCompareImages(token, uris);
      setResult(res);
    } catch (e) {
      setError(e instanceof Error ? e.message : '对比失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <SafeAreaView style={styles.root} edges={['top']}>
      <ScrollView contentContainerStyle={styles.scroll}>
        <ThemedText type="title">图片质量判断</ThemedText>
        <ThemedText style={styles.hint}>上传多张候选照片，AI 对比并推荐最佳一张</ThemedText>

        {!token ? (
          <ThemedView style={styles.card}>
            <ThemedText type="subtitle">登录</ThemedText>
            <TextInput
              style={styles.input}
              placeholder="用户名或邮箱"
              placeholderTextColor="#888"
              value={account}
              onChangeText={setAccount}
              autoCapitalize="none"
            />
            <TextInput
              style={styles.input}
              placeholder="密码"
              placeholderTextColor="#888"
              secureTextEntry
              value={password}
              onChangeText={setPassword}
            />
            <Pressable style={styles.btn} onPress={handleLogin} disabled={loading}>
              <ThemedText style={styles.btnText}>登录</ThemedText>
            </Pressable>
          </ThemedView>
        ) : (
          <ThemedText style={styles.ok}>已登录 ✓</ThemedText>
        )}

        <ThemedView style={styles.card}>
          <Pressable style={styles.btnSecondary} onPress={pickImages}>
            <ThemedText>选择 2–8 张照片</ThemedText>
          </Pressable>
          {uris.length > 0 && (
            <ThemedText style={styles.meta}>已选 {uris.length} 张</ThemedText>
          )}
          <View style={styles.thumbs}>
            {uris.map((uri, i) => (
              <Image key={uri} source={{ uri }} style={styles.thumb} />
            ))}
          </View>
          <Pressable style={styles.btn} onPress={handleCompare} disabled={loading}>
            {loading ? (
              <ActivityIndicator color="#0A0A0A" />
            ) : (
              <ThemedText style={styles.btnText}>开始智能选片</ThemedText>
            )}
          </Pressable>
        </ThemedView>

        {error ? <ThemedText style={styles.err}>{error}</ThemedText> : null}

        {result ? (
          <ThemedView style={styles.card}>
            <ThemedText type="subtitle">最佳推荐</ThemedText>
            <ThemedText style={styles.best}>
              第 {result.best_index + 1} 张 · 综合最高
            </ThemedText>
            <ThemedText>{result.best_reason}</ThemedText>
            <ThemedText style={styles.summary}>{result.summary}</ThemedText>

            <ThemedText type="subtitle" style={styles.section}>
              排序结果
            </ThemedText>
            {result.photos.map((p, rank) => (
              <View key={p.index} style={styles.rankRow}>
                <ThemedText style={styles.rankNum}>#{rank + 1}</ThemedText>
                <View style={styles.rankBody}>
                  <ThemedText>
                    原图序号 {p.index + 1} · 综合 {p.overall_score} 分
                  </ThemedText>
                  <ThemedText style={styles.dim}>
                    构图{p.dimension_scores.composition} 色彩{p.dimension_scores.color}{' '}
                    曝光{p.dimension_scores.exposure} 清晰{p.dimension_scores.sharpness}{' '}
                    创意{p.dimension_scores.creativity}
                  </ThemedText>
                  <ThemedText style={styles.pro}>优点：{p.pros}</ThemedText>
                  <ThemedText style={styles.con}>不足：{p.cons}</ThemedText>
                </View>
              </View>
            ))}
          </ThemedView>
        ) : null}
      </ScrollView>
      <TabBar />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#0A0A0A' },
  scroll: { padding: 20, paddingBottom: 100, gap: 12 },
  hint: { color: '#999', marginBottom: 8 },
  card: { padding: 16, borderRadius: 8, gap: 10 },
  input: {
    borderWidth: 1,
    borderColor: '#333',
    borderRadius: 6,
    padding: 10,
    color: '#fff',
  },
  btn: {
    backgroundColor: '#fff',
    padding: 12,
    borderRadius: 6,
    alignItems: 'center',
  },
  btnSecondary: {
    borderWidth: 1,
    borderColor: '#444',
    padding: 12,
    borderRadius: 6,
    alignItems: 'center',
  },
  btnText: { color: '#0A0A0A', fontWeight: '600' },
  ok: { color: '#6c6' },
  meta: { color: '#888', fontSize: 13 },
  thumbs: { flexDirection: 'row', flexWrap: 'wrap', gap: 8 },
  thumb: { width: 72, height: 72, borderRadius: 4 },
  err: { color: '#f66' },
  best: { fontSize: 18, fontWeight: '700', marginVertical: 6 },
  summary: { color: '#aaa', marginTop: 8 },
  section: { marginTop: 16 },
  rankRow: { flexDirection: 'row', gap: 10, marginTop: 12 },
  rankNum: { color: '#fff', fontWeight: '700', width: 28 },
  rankBody: { flex: 1, gap: 4 },
  dim: { color: '#888', fontSize: 12 },
  pro: { color: '#9c9', fontSize: 13 },
  con: { color: '#c99', fontSize: 13 },
});
