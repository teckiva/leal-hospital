import React from 'react';
import './OPDTemplate.css';

const OPDTemplate = ({ patientData, opdData }) => {
  // Format date for display
  const formatDate = (dateString) => {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-IN');
  };

  // Format medicines for display
  const formatMedicines = (medicines) => {
    if (!medicines || !Array.isArray(medicines)) return [];
    return medicines;
  };

  return (
    <div className="opd-container">
      {/* Header Section */}
      <div className="opd-header">
        <div className="medical-symbol">
          <svg width="90" height="110" viewBox="0 0 90 110" fill="none" xmlns="http://www.w3.org/2000/svg">
            <defs>
              {/* Gradients for 3D effect */}
              <radialGradient id="ballGradient" cx="30%" cy="30%">
                <stop offset="0%" stopColor="#FFFEF7" />
                <stop offset="40%" stopColor="#F5F0DC" />
                <stop offset="100%" stopColor="#D4C5A0" />
              </radialGradient>

              <linearGradient id="staffGradient" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#D4C5A0" />
                <stop offset="50%" stopColor="#F5F0DC" />
                <stop offset="100%" stopColor="#C0B090" />
              </linearGradient>

              <linearGradient id="wingGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" stopColor="#F8F4E6" />
                <stop offset="50%" stopColor="#E8DCC0" />
                <stop offset="100%" stopColor="#D4C5A0" />
              </linearGradient>

              <linearGradient id="snakeGradient" x1="30%" y1="30%" x2="70%" y2="70%">
                <stop offset="0%" stopColor="#FFFEF7" />
                <stop offset="50%" stopColor="#F0E8D0" />
                <stop offset="100%" stopColor="#D0C0A0" />
              </linearGradient>

              {/* Shadow filters */}
              <filter id="shadow" x="-50%" y="-50%" width="200%" height="200%">
                <feGaussianBlur in="SourceAlpha" stdDeviation="2"/>
                <feOffset dx="2" dy="3" result="offsetblur"/>
                <feComponentTransfer>
                  <feFuncA type="linear" slope="0.3"/>
                </feComponentTransfer>
                <feMerge>
                  <feMergeNode/>
                  <feMergeNode in="SourceGraphic"/>
                </feMerge>
              </filter>
            </defs>

            {/* Shadow layer */}
            <g opacity="0.2" transform="translate(3, 4)">
              <ellipse cx="45" cy="8" rx="7" ry="7" fill="#000"/>
              <rect x="40" y="15" width="10" height="75" fill="#000"/>
              <path d="M35 25 Q25 35 20 50 L15 75 L25 80 L35 50 Z" fill="#000"/>
              <path d="M55 25 Q65 35 70 50 L75 75 L65 80 L55 50 Z" fill="#000"/>
            </g>

            {/* Main Caduceus */}
            <g filter="url(#shadow)">
              {/* Top ball */}
              <ellipse cx="45" cy="8" rx="7" ry="7" fill="url(#ballGradient)" stroke="#B8A882" strokeWidth="0.5"/>
              <ellipse cx="43" cy="6" rx="2" ry="2" fill="#FFFFFF" opacity="0.6"/>

              {/* Central staff */}
              <rect x="40" y="15" width="10" height="75" fill="url(#staffGradient)" stroke="#B8A882" strokeWidth="0.5"/>
              <rect x="41" y="15" width="3" height="75" fill="#FFFFFF" opacity="0.15"/>

              {/* Left wing */}
              <path d="M35 22 Q25 25 18 32 Q15 38 15 45 Q15 48 18 50 L25 48 Q28 40 35 35 Z"
                    fill="url(#wingGradient)" stroke="#B8A882" strokeWidth="0.8"/>
              <path d="M35 24 Q28 27 23 32 Q22 35 23 38 L27 37 Q29 33 35 30 Z"
                    fill="#FFFFFF" opacity="0.25"/>
              <line x1="23" y1="35" x2="18" y2="38" stroke="#B8A882" strokeWidth="1.5"/>
              <line x1="26" y1="32" x2="21" y2="35" stroke="#B8A882" strokeWidth="1.5"/>
              <line x1="29" y1="29" x2="24" y2="32" stroke="#B8A882" strokeWidth="1.5"/>

              {/* Right wing */}
              <path d="M55 22 Q65 25 72 32 Q75 38 75 45 Q75 48 72 50 L65 48 Q62 40 55 35 Z"
                    fill="url(#wingGradient)" stroke="#B8A882" strokeWidth="0.8"/>
              <path d="M55 24 Q62 27 67 32 Q68 35 67 38 L63 37 Q61 33 55 30 Z"
                    fill="#FFFFFF" opacity="0.25"/>
              <line x1="67" y1="35" x2="72" y2="38" stroke="#B8A882" strokeWidth="1.5"/>
              <line x1="64" y1="32" x2="69" y2="35" stroke="#B8A882" strokeWidth="1.5"/>
              <line x1="61" y1="29" x2="66" y2="32" stroke="#B8A882" strokeWidth="1.5"/>

              {/* Left snake (spiral around staff) */}
              <path d="M35 25 Q30 30 28 38 Q27 48 30 58 Q33 68 35 75"
                    stroke="url(#snakeGradient)" strokeWidth="4" fill="none" strokeLinecap="round"/>
              <path d="M35 25 Q32 30 31 38 Q30 48 32 58 Q34 68 35 75"
                    stroke="#FFFFFF" strokeWidth="1.5" fill="none" opacity="0.4" strokeLinecap="round"/>
              <ellipse cx="35" cy="25" rx="3" ry="3" fill="url(#snakeGradient)"/>
              <ellipse cx="33" cy="24" rx="1" ry="1" fill="#FFFFFF" opacity="0.6"/>
              <ellipse cx="35" cy="75" rx="2.5" ry="2.5" fill="url(#snakeGradient)"/>

              {/* Right snake (spiral around staff) */}
              <path d="M55 25 Q60 30 62 38 Q63 48 60 58 Q57 68 55 75"
                    stroke="url(#snakeGradient)" strokeWidth="4" fill="none" strokeLinecap="round"/>
              <path d="M55 25 Q58 30 59 38 Q60 48 58 58 Q56 68 55 75"
                    stroke="#FFFFFF" strokeWidth="1.5" fill="none" opacity="0.4" strokeLinecap="round"/>
              <ellipse cx="55" cy="25" rx="3" ry="3" fill="url(#snakeGradient)"/>
              <ellipse cx="57" cy="24" rx="1" ry="1" fill="#FFFFFF" opacity="0.6"/>
              <ellipse cx="55" cy="75" rx="2.5" ry="2.5" fill="url(#snakeGradient)"/>

              {/* Bottom base */}
              <ellipse cx="45" cy="90" rx="8" ry="4" fill="url(#ballGradient)" stroke="#B8A882" strokeWidth="0.5"/>
              <ellipse cx="45" cy="89" rx="6" ry="3" fill="#FFFFFF" opacity="0.2"/>
            </g>
          </svg>
        </div>

        <div className="hospital-info">
          <h1 className="hospital-name">LAEL ENT HOSPITAL</h1>
          <p className="hospital-address-hindi">
            पता वार्ड नं 7, स्वामी विवेकानन्द नगर पहरौना, ( निकट अम्बे चौक), कुशीनगर
          </p>
          <p className="hospital-contact">मो.7080920551, 9129272551</p>
        </div>

        <div className="hospital-accreditation">
          <p>पूर्व रेजिडेन्ट सफदरजंग अस्पताल, नई दिल्ली</p>
          <p>पूर्व रेजिडेन्ट KGMU लखनऊ</p>
          <p>पूर्व वरिष्ठ ENT चिकित्सक, संयुक्त जिला चिकित्सालय</p>
          <p className="location-line">पहरौना, कुशीनगर</p>
          <p>पूर्व असिस्टेन्ट प्रोफेस (Assistant Professor) मेडिकल कालेज, देवरिया</p>
        </div>
      </div>

      <div className="doctor-name-section">
        <h2 className="doctor-name">डा. संतोष कुमार</h2>
        <p className="doctor-qualifications">MBBS,DLO,MS ENT (KGMU Lucknow)</p>
        <p className="doctor-specialization">कान, नाक, गला एवं हैड नेक सर्जन</p>
        <p className="doctor-registration">रजिस्ट्रेशन नं 48357</p>
      </div>

      {/* Patient Details Section - With Data Binding */}
      <div className="patient-details-section">
        <div className="patient-line-1">
          <span className="label">Pt. Name</span>
          <span className="filled-data">{patientData?.name || ''}</span>
          <span className="dotted-line"></span>
          <span className="label">Age</span>
          <span className="filled-data">{patientData?.age || ''}</span>
          <span className="dotted-line-short"></span>
          <span className="label">Sex</span>
          <span className="filled-data">{patientData?.sex || ''}</span>
          <span className="dotted-line-short"></span>
        </div>
        <div className="patient-line-2">
          <span className="label">Address</span>
          <span className="filled-data">
            {patientData?.address
              ? `${patientData.address.locality || ''}, ${patientData.address.city || ''}, ${patientData.address.state || ''} ${patientData.address.pincode || ''}`
              : ''}
          </span>
          <span className="dotted-line-long"></span>
          <span className="label">Date</span>
          <span className="filled-data">{formatDate(patientData?.date)}</span>
          <span className="dotted-line-short"></span>
        </div>
      </div>

      {/* Main Content Section */}
      <div className="opd-main-content">
        {/* Left Sidebar - Vitals */}
        <div className="left-sidebar">
          <div className="vitals-item">
            <strong>BP</strong>
            <div className="vitals-value">{opdData?.vitals?.bp || ''}</div>
          </div>
          <div className="vitals-item">
            <strong>Pulse</strong>
            <div className="vitals-value">{opdData?.vitals?.pulse || ''}</div>
          </div>
          <div className="vitals-item">
            <strong>Weight</strong>
            <div className="vitals-value">{opdData?.vitals?.weight || ''}</div>
          </div>
          <div className="vitals-item">
            <strong>RBS</strong>
            <div className="vitals-value">{opdData?.vitals?.rbs || ''}</div>
          </div>
          <div className="ear-examination">
            <strong>Right ear</strong>
            <div className="ear-diagram">
              <svg width="80" height="80" viewBox="0 0 80 80">
                <circle cx="40" cy="40" r="35" fill="white" stroke="black" strokeWidth="2" />
                <line x1="40" y1="40" x2="60" y2="30" stroke="black" strokeWidth="2" />
                <circle cx="40" cy="40" r="3" fill="black" />
              </svg>
            </div>
            <div className="ear-notes">{opdData?.examination?.rightEar || ''}</div>
          </div>
          <div className="ear-examination">
            <strong>Left ear</strong>
            <div className="ear-diagram">
              <svg width="80" height="80" viewBox="0 0 80 80">
                <circle cx="40" cy="40" r="35" fill="white" stroke="black" strokeWidth="2" />
                <line x1="40" y1="40" x2="20" y2="30" stroke="black" strokeWidth="2" />
                <circle cx="40" cy="40" r="3" fill="black" />
              </svg>
            </div>
            <div className="ear-notes">{opdData?.examination?.leftEar || ''}</div>
          </div>
          <div className="vitals-item">
            <strong>AR</strong>
            <div className="vitals-value">{opdData?.examination?.ar || ''}</div>
          </div>
          <div className="vitals-item">
            <strong>OC/OP</strong>
            <div className="vitals-value">{opdData?.examination?.ocOp || ''}</div>
          </div>
          <div className="findings-section">
            <strong>VDL FINDINGS</strong>
            <div className="findings-value">{opdData?.examination?.vdlFindings || ''}</div>
          </div>
          <div className="findings-section">
            <strong>STATUS OF NECK</strong>
            <div className="findings-value">{opdData?.examination?.statusOfNeck || ''}</div>
          </div>
          <div className="findings-section">
            <strong>NODES</strong>
            <div className="findings-value">{opdData?.examination?.nodes || ''}</div>
          </div>
        </div>

        {/* Center Content - THE 3 INVISIBLE PARTS */}
        <div className="center-content">
          {/* Rx Symbol */}
          <div className="rx-symbol">R<sub>x</sub></div>

          {/* Medical Symbol Watermark */}
          <div className="caduceus-watermark">
            <svg width="300" height="500" viewBox="0 0 300 500" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M150 50 L140 100 L150 120 L160 100 Z" fill="rgba(218, 165, 32, 0.15)" />
              <ellipse cx="150" cy="50" rx="20" ry="20" fill="rgba(218, 165, 32, 0.15)" />
              <path d="M130 120 Q100 140 80 160 L40 300 Q60 320 80 320 L120 180 Z" fill="rgba(218, 165, 32, 0.15)" />
              <path d="M170 120 Q200 140 220 160 L260 300 Q240 320 220 320 L180 180 Z" fill="rgba(218, 165, 32, 0.15)" />
              <rect x="135" y="120" width="30" height="300" fill="rgba(218, 165, 32, 0.15)" />
              <ellipse cx="150" cy="420" rx="30" ry="20" fill="rgba(218, 165, 32, 0.15)" />
            </svg>
          </div>

          {/* PART 1: SYMPTOMS (Invisible/Print-only section) */}
          <div className="prescription-section symptoms-section">
            <div className="section-title">SYMPTOMS:</div>
            <div className="section-content">
              {opdData?.symptoms && Array.isArray(opdData.symptoms) && opdData.symptoms.length > 0 ? (
                <ul className="symptoms-list">
                  {opdData.symptoms.map((symptom, index) => (
                    <li key={index}>{symptom}</li>
                  ))}
                </ul>
              ) : (
                <div className="empty-space"></div>
              )}
            </div>
          </div>

          {/* PART 2: PRESCRIPTION & MEDICINES (Invisible/Print-only section) */}
          <div className="prescription-section medicines-section">
            <div className="section-title">PRESCRIPTION:</div>
            <div className="section-content">
              {opdData?.prescription && Array.isArray(opdData.prescription) && opdData.prescription.length > 0 ? (
                <ul className="prescription-list">
                  {opdData.prescription.map((item, index) => (
                    <li key={index}>{item}</li>
                  ))}
                </ul>
              ) : (
                <div className="empty-space"></div>
              )}
            </div>

            {/* Medicines Table */}
            {opdData?.medicines && Array.isArray(opdData.medicines) && opdData.medicines.length > 0 && (
              <div className="medicines-table">
                <table>
                  <thead>
                    <tr>
                      <th>Medicine</th>
                      <th>Volume</th>
                      <th>Morning</th>
                      <th>Afternoon</th>
                      <th>Night</th>
                    </tr>
                  </thead>
                  <tbody>
                    {formatMedicines(opdData.medicines).map((medicine, index) => (
                      <tr key={index}>
                        <td>{medicine.name || ''}</td>
                        <td>{medicine.volume || ''}</td>
                        <td>{medicine.timings?.morning ? '✓' : ''}</td>
                        <td>{medicine.timings?.afternoon ? '✓' : ''}</td>
                        <td>{medicine.timings?.night ? '✓' : ''}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          {/* PART 3: FUTURE SUGGESTIONS (Invisible/Print-only section) */}
          <div className="prescription-section future-section">
            <div className="section-title">FUTURE SUGGESTIONS:</div>
            <div className="section-content">
              {opdData?.futureSuggestion && Array.isArray(opdData.futureSuggestion) && opdData.futureSuggestion.length > 0 ? (
                <ul className="future-list">
                  {opdData.futureSuggestion.map((suggestion, index) => (
                    <li key={index}>{suggestion}</li>
                  ))}
                </ul>
              ) : (
                <div className="empty-space"></div>
              )}
            </div>
          </div>

          {/* Legal Text */}
          <div className="legal-text">Not For Medico Legal Purpose</div>

          {/* Timing Section */}
          <div className="timing-section">
            <p className="opd-timing">ओपीडी समय सोमवार से शनिवार 11 AM - 3 PM ; रविवार बन्दी</p>
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="opd-footer"></div>
    </div>
  );
};

export default OPDTemplate;
