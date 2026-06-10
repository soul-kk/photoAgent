import React from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  SafeAreaView,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  ActivityIndicator,
} from 'react-native';
import { useRouter } from 'expo-router';
import { useAuth } from '@/context/auth';

type Mode = 'login' | 'register';

export default function AuthScreen() {
  const router = useRouter();
  const { login, register } = useAuth();

  const [mode, setMode] = React.useState<Mode>('login');
  const [account, setAccount] = React.useState('');
  const [username, setUsername] = React.useState('');
  const [email, setEmail] = React.useState('');
  const [password, setPassword] = React.useState('');
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState('');

  function switchMode(next: Mode) {
    setMode(next);
    setError('');
    setAccount('');
    setUsername('');
    setEmail('');
    setPassword('');
  }

  async function handleSubmit() {
    setError('');
    if (mode === 'login') {
      if (!account.trim() || !password.trim()) {
        setError('请填写账号和密码');
        return;
      }
    } else {
      if (!username.trim() || !email.trim() || !password.trim()) {
        setError('请填写所有字段');
        return;
      }
    }

    setLoading(true);
    try {
      if (mode === 'login') {
        await login(account.trim(), password);
      } else {
        await register(username.trim(), email.trim(), password);
      }
      router.replace('/');
    } catch (e: any) {
      setError(e.message ?? '操作失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  }

  return (
    <SafeAreaView style={styles.root}>
      <KeyboardAvoidingView
        style={styles.flex}
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      >
        <ScrollView
          contentContainerStyle={styles.scroll}
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}
        >
          {/* Logo */}
          <View style={styles.logo}>
            <Text style={styles.logoTitle}>LENS AI</Text>
            <Text style={styles.logoSub}>Photography Intelligence</Text>
          </View>

          {/* Mode tabs */}
          <View style={styles.modePill}>
            {(['login', 'register'] as Mode[]).map((m) => (
              <TouchableOpacity
                key={m}
                style={[styles.modeTab, mode === m && styles.modeTabActive]}
                onPress={() => switchMode(m)}
                activeOpacity={0.8}
              >
                <Text style={[styles.modeTabText, mode === m && styles.modeTabTextActive]}>
                  {m === 'login' ? '登录' : '注册'}
                </Text>
              </TouchableOpacity>
            ))}
          </View>

          {/* Form */}
          <View style={styles.form}>
            {mode === 'register' && (
              <FormField
                label="用户名"
                value={username}
                onChangeText={setUsername}
                placeholder="请输入用户名"
                autoCapitalize="none"
              />
            )}
            <FormField
              label={mode === 'login' ? '账号' : '邮箱'}
              value={mode === 'login' ? account : email}
              onChangeText={mode === 'login' ? setAccount : setEmail}
              placeholder={mode === 'login' ? '用户名或邮箱' : '请输入邮箱'}
              keyboardType={mode === 'register' ? 'email-address' : 'default'}
              autoCapitalize="none"
            />
            <FormField
              label="密码"
              value={password}
              onChangeText={setPassword}
              placeholder="请输入密码"
              secureTextEntry
            />
          </View>

          {/* Error */}
          {error ? <Text style={styles.error}>{error}</Text> : null}

          {/* Submit */}
          <TouchableOpacity
            style={[styles.submitBtn, loading && styles.submitBtnLoading]}
            onPress={handleSubmit}
            activeOpacity={0.85}
            disabled={loading}
          >
            {loading ? (
              <ActivityIndicator color="#0A0A0A" size="small" />
            ) : (
              <Text style={styles.submitBtnText}>
                {mode === 'login' ? '登录' : '注册'}
              </Text>
            )}
          </TouchableOpacity>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

// ─── FormField sub-component ──────────────────────────────────────────────────

function FormField({
  label,
  value,
  onChangeText,
  placeholder,
  secureTextEntry,
  keyboardType,
  autoCapitalize,
}: {
  label: string;
  value: string;
  onChangeText: (v: string) => void;
  placeholder?: string;
  secureTextEntry?: boolean;
  keyboardType?: any;
  autoCapitalize?: any;
}) {
  const [focused, setFocused] = React.useState(false);
  return (
    <View style={fieldStyles.card}>
      <Text style={fieldStyles.label}>{label}</Text>
      <TextInput
        style={fieldStyles.input}
        value={value}
        onChangeText={onChangeText}
        placeholder={placeholder}
        placeholderTextColor="#444444"
        secureTextEntry={secureTextEntry}
        keyboardType={keyboardType}
        autoCapitalize={autoCapitalize ?? 'none'}
        onFocus={() => setFocused(true)}
        onBlur={() => setFocused(false)}
        returnKeyType="next"
      />
      <View style={[fieldStyles.underline, (focused || value.length > 0) && fieldStyles.underlineActive]} />
    </View>
  );
}

const fieldStyles = StyleSheet.create({
  card: {
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    padding: 18,
    gap: 10,
  },
  label: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 10,
    color: '#555555',
    letterSpacing: 2,
  },
  input: {
    fontFamily: 'DMSans_500Medium',
    fontSize: 15,
    color: '#FFFFFF',
    paddingVertical: 0,
  },
  underline: { height: 1, backgroundColor: '#2A2A2A' },
  underlineActive: { backgroundColor: '#FFFFFF' },
});

// ─── Styles ───────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#0A0A0A' },
  flex: { flex: 1 },
  scroll: {
    flexGrow: 1,
    paddingHorizontal: 24,
    paddingTop: 60,
    paddingBottom: 48,
    gap: 20,
  },
  logo: { alignItems: 'center', gap: 8, marginBottom: 16 },
  logoTitle: {
    fontFamily: 'PlayfairDisplay_700Bold',
    fontSize: 48,
    color: '#FFFFFF',
    letterSpacing: 6,
  },
  logoSub: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 11,
    color: '#555555',
    letterSpacing: 3,
  },
  modePill: {
    flexDirection: 'row',
    backgroundColor: '#141414',
    borderRadius: 36,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    padding: 4,
  },
  modeTab: {
    flex: 1,
    height: 40,
    alignItems: 'center',
    justifyContent: 'center',
    borderRadius: 26,
  },
  modeTabActive: { backgroundColor: '#FFFFFF' },
  modeTabText: {
    fontFamily: 'DMMono_500Medium',
    fontSize: 13,
    color: '#555555',
    letterSpacing: 1,
  },
  modeTabTextActive: { color: '#0A0A0A' },
  form: { gap: 12 },
  error: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 13,
    color: '#FF5555',
    textAlign: 'center',
  },
  submitBtn: {
    height: 52,
    backgroundColor: '#FFFFFF',
    borderRadius: 4,
    alignItems: 'center',
    justifyContent: 'center',
    marginTop: 4,
  },
  submitBtnLoading: { opacity: 0.7 },
  submitBtnText: {
    fontFamily: 'DMSans_700Bold',
    fontSize: 15,
    color: '#0A0A0A',
  },
});
