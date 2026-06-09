const PEM_BLOCK_PATTERN = /-----BEGIN ([^-]+)-----[\s\S]*?-----END \1-----/g;
const CERT_TYPE = 'CERTIFICATE';
const KEY_TYPES = ['PRIVATE KEY', 'RSA PRIVATE KEY', 'EC PRIVATE KEY'];

export function parsePEMBlocks(value = '') {
  const blocks = [];
  const text = String(value || '');
  let match;

  while ((match = PEM_BLOCK_PATTERN.exec(text)) !== null) {
    blocks.push({
      type: match[1],
      pem: match[0].trim() + '\n'
    });
  }

  return blocks;
}

export function splitTLSPEM(value = '') {
  const blocks = parsePEMBlocks(value);
  const certificates = blocks.filter((block) => block.type === CERT_TYPE).map((block) => block.pem);
  const privateKeys = blocks.filter((block) => KEY_TYPES.includes(block.type)).map((block) => block.pem);

  return {
    cert: certificates.join(''),
    pkey: privateKeys.join(''),
    certificateCount: certificates.length,
    privateKeyCount: privateKeys.length,
  };
}

export function summarizeTLSPEM({ cert = '', pkey = '', cacert = '' } = {}) {
  const certParts = splitTLSPEM(cert);
  const keyParts = splitTLSPEM(pkey);
  const caParts = splitTLSPEM(cacert);

  return {
    certificateCount: certParts.certificateCount,
    hasPrivateKey: Boolean(keyParts.privateKeyCount),
    caCertificateCount: caParts.certificateCount,
  };
}
