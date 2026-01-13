import { format, formatInTimeZone } from 'date-fns-tz';
import { th } from 'date-fns/locale';

const THAILAND_TIMEZONE = 'Asia/Bangkok';

export const formatThailandTime = (date: Date | string, formatStr: string = 'yyyy-MM-dd HH:mm:ss') => {
  const dateObj = typeof date === 'string' ? new Date(date) : date;
  return formatInTimeZone(dateObj, THAILAND_TIMEZONE, formatStr, { locale: th });
};

export const getThailandDate = (date?: Date | string) => {
  if (!date) return new Date();
  const dateObj = typeof date === 'string' ? new Date(date) : date;
  return formatInTimeZone(dateObj, THAILAND_TIMEZONE, 'yyyy-MM-dd');
};

export const getThailandNow = () => {
  return new Date(new Date().toLocaleString('en-US', { timeZone: THAILAND_TIMEZONE }));
};






