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
import { fetchToneStyle, login, type ToneStyleResponse } from '@/lib/api';

const PARAM_KEYS = [
  ['exposure', '曝光'],
  ['contrast', '对比度'],
  ['highlights', '高光'],
  ['shadows', '阴影'],
  ['saturation', '饱和度'],
  ['vibrance', '自然饱和'],
  ['temperature', '色温'],
  ['tint', '色调'],
] as const;

export default function ToneStyleScreen() {
  const [token, setToken] = useState('');
  const [account, setAccount] = useState('');
  const [password, setPassword] = useState('');
  const [styleDesc, setStyleDesc] = useState('');
  const [imageUri, setImageUri] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<ToneStyleResponse | null>(null);

  const pickImage = async () => {
    const perm = await ImagePicker.requestMediaLibraryPermissionsAsync();
    if (!perm.granted) {
      setError('需要相册权限');
      return;
    }
    const picked = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ['images'],
      quality: 0.85,
    });
    if (!picked.canceled && picked.assets[0]) {
      setImageUri(picked.assets[0].uri);
      setError(null);
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

  const handleSubmit = async () => {
    if (!token) {
      setError('请先登录');
      return;
    }
    if (!styleDesc.trim()) {
      setError('请描述想要的影调风格');
      return;
    }
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      const res = await fetchToneStyle(token, styleDesc.trim(), imageUri);
      setResult(res);
    } catch (e) {
      setError(e instanceof Error ? e.message : '获取建议失败');
    } finally {
      setLoading(false);
    }
  };

  const style = result?.style;

  return (
    <SafeAreaView style={styles.root} edges={['top']}>
      <ScrollView contentContainerStyle={styles.scroll}>
        <ThemedText type="title">影调风格建议</ThemedText>
        <ThemedText style={styles.hint}>
          描述目标风格（如「王家卫电影感」「富士清冷色系」），可选上传参考图
        </ThemedText>

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
          <TextInput
            style={[styles.input, styles.textarea]}
            placeholder="例如：王家卫电影感、偏黄绿、低饱和"
            placeholderTextColor="#888"
            value={styleDesc}
            onChangeText={setStyleDesc}
            multiline
          />
          <Pressable style={styles.btnSecondary} onPress={pickImage}>
            <ThemedText>{imageUri ? '更换参考图' : '添加参考图（可选）'}</ThemedText>
          </Pressable>
          {imageUri ? <Image source={{ uri: imageUri }} style={styles.preview} /> : null}
          <Pressable style={styles.btn} onPress={handleSubmit} disabled={loading}>
            {loading ? (
              <ActivityIndicator color="#0A0A0A" />
            ) : (
              <ThemedText style={styles.btnText}>生成后期参数</ThemedText>
            )}
          </Pressable>
        </ThemedView>

        {error ? <ThemedText style={styles.err}>{error}</ThemedText> : null}

        {style ? (
          <ThemedView style={styles.card}>
            <ThemedText type="subtitle">{style.style_name}</ThemedText>
            <ThemedText style={styles.match}>{style.style_match_summary}</ThemedText>

            <ThemedText type="subtitle" style={styles.section}>
              后期参数
            </ThemedText>
            {PARAM_KEYS.map(([key, label]) => (
              <View key={key} style={styles.paramRow}>
                <ThemedText style={styles.paramLabel}>{label}</ThemedText>
                <ThemedText style={styles.paramVal}>
                  {style.parameters[key] || '—'}
                </ThemedText>
              </View>
            ))}

            <ThemedText type="subtitle" style={styles.section}>
              操作步骤
            </ThemedText>
            {style.adjustment_notes.map((n, i) => (
              <ThemedText key={i} style={styles.note}>
                {i + 1}. {n}
              </ThemedText>
            ))}

            <ThemedText type="subtitle" style={styles.section}>
              前后对比示意
            </ThemedText>
            <ThemedText style={styles.block}>{style.before_after_description}</ThemedText>

            <ThemedText type="subtitle" style={styles.section}>
              预览提示
            </ThemedText>
            <ThemedText style={styles.block}>{style.preview_hints}</ThemedText>
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
  textarea: { minHeight: 80, textAlignVertical: 'top' },
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
  preview: { width: '100%', height: 160, borderRadius: 6 },
  err: { color: '#f66' },
  match: { color: '#aaa', marginTop: 4 },
  section: { marginTop: 14 },
  paramRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: 6,
    borderBottomWidth: 1,
    borderBottomColor: '#222',
  },
  paramLabel: { color: '#888' },
  paramVal: { color: '#fff', maxWidth: '55%', textAlign: 'right' },
  note: { color: '#ccc', marginTop: 6, lineHeight: 20 },
  block: { color: '#bbb', lineHeight: 22, marginTop: 4 },
});
