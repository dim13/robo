package robo

// found in firmware V2.30
var easteregg = []string{
	`FU5440,4000`,
	`FM1`,
	`TB50,0`,
	`FO5440`,
	`&100,100,100`,
	`\0,0`,
	`Z5940,4200`,
	`L100,1,40,80`,
	`^0,100`,

	`M44.006,1143.797`,
	`D1045.895,1143.797`,

	`M1045.895,1992.228`,
	`D44.006,1992.228`,

	`M2679.256,1142.033`,
	`D2679.256,1992.228`,

	`M2681.02,3009.992`,
	`D2681.02,1992.228`,
	`D1045.895,1992.228`,
	`D1045.895,1142.033`,
	`D2681.02,1142.033`,
	`D2681.02,124.269`,

	`L0`,

	`M29.895,1143.797`,
	`D59.881,1073.242`,
	`D1015.909,1073.242`,
	`D1045.895,1142.033`,
	`D1047.659,1142.033`,
	`D1047.659,126.033`,
	`D1345.756,126.033`,
	`D1345.756,124.269`,
	`D1486.867,124.269`,
	`D1486.867,124.269`,
	`D1499.214,101.339`,
	`D1513.326,80.172`,
	`D1532.728,62.533`,
	`D1552.131,46.658`,
	`D1575.062,34.311`,
	`D1597.992,23.728`,
	`D1624.451,18.436`,
	`D1650.909,16.672`,
	`D1650.909,16.672`,
	`D1679.131,18.436`,
	`D1703.826,23.728`,
	`D1728.52,34.311`,
	`D1751.451,46.658`,
	`D1770.853,62.533`,
	`D1788.492,80.172`,
	`D1804.367,101.339`,
	`D1814.951,124.269`,
	`D1880.214,124.269`,
	`D1880.214,124.269`,
	`D1892.562,101.339`,
	`D1906.673,80.172`,
	`D1924.312,62.533`,
	`D1945.478,46.658`,
	`D1966.645,34.311`,
	`D1991.339,23.728`,
	`D2017.798,18.436`,
	`D2044.256,16.672`,
	`D2044.256,16.672`,
	`D2070.714,18.436`,
	`D2097.173,23.728`,
	`D2121.867,34.311`,
	`D2143.034,46.658`,
	`D2164.201,62.533`,
	`D2181.839,80.172`,
	`D2195.951,101.339`,
	`D2208.298,124.269`,
	`D2337.062,124.269`,
	`D2337.062,126.033`,
	`D2681.02,126.033`,
	`D2681.02,113.686`,
	`D2813.312,170.131`,
	`D2813.312,1039.728`,
	`D2681.02,1094.408`,
	`D2681.02,1143.797`,
	`D2811.548,1143.797`,
	`D2811.548,1992.228`,
	`D2681.02,1992.228`,
	`D2681.02,2032.797`,
	`D2813.312,2089.242`,
	`D2813.312,2957.075`,
	`D2681.02,3013.519`,
	`D2681.02,3009.992`,
	`D2322.951,3009.992`,
	`D2322.951,3009.992`,
	`D2307.076,3009.992`,
	`D2294.728,3008.228`,
	`D2213.589,3008.228`,
	`D2213.589,3008.228`,
	`D2210.062,3020.575`,
	`D2203.006,3032.922`,
	`D2197.714,3045.27`,
	`D2188.895,3057.617`,
	`D2180.076,3068.2`,
	`D2171.256,3078.783`,
	`D2160.673,3087.603`,
	`D2150.089,3096.422`,
	`D2139.506,3105.242`,
	`D2127.159,3112.297`,
	`D2114.812,3117.589`,
	`D2100.701,3122.881`,
	`D2086.59,3126.408`,
	`D2072.478,3129.936`,
	`D2058.367,3131.7`,
	`D2044.256,3131.7`,
	`D2044.256,3131.7`,
	`D2028.381,3131.7`,
	`D2014.27,3129.936`,
	`D2000.159,3126.408`,
	`D1987.812,3122.881`,
	`D1973.701,3117.589`,
	`D1961.353,3112.297`,
	`D1949.006,3105.242`,
	`D1938.423,3096.422`,
	`D1926.076,3087.603`,
	`D1917.256,3078.783`,
	`D1906.673,3068.2`,
	`D1899.617,3057.617`,
	`D1890.798,3045.27`,
	`D1883.742,3032.922`,
	`D1878.451,3020.575`,
	`D1873.159,3008.228`,
	`D1822.006,3008.228`,
	`D1822.006,3008.228`,
	`D1816.714,3020.575`,
	`D1811.423,3032.922`,
	`D1804.367,3045.27`,
	`D1797.312,3057.617`,
	`D1788.492,3068.2`,
	`D1779.673,3078.783`,
	`D1769.09,3087.603`,
	`D1758.506,3096.422`,
	`D1746.159,3105.242`,
	`D1733.812,3112.297`,
	`D1721.464,3117.589`,
	`D1709.117,3122.881`,
	`D1695.006,3126.408`,
	`D1680.895,3129.936`,
	`D1666.784,3131.7`,
	`D1650.909,3131.7`,
	`D1650.909,3131.7`,
	`D1636.798,3131.7`,
	`D1622.687,3129.936`,
	`D1608.576,3126.408`,
	`D1594.465,3122.881`,
	`D1582.117,3117.589`,
	`D1568.006,3112.297`,
	`D1555.659,3105.242`,
	`D1545.076,3096.422`,
	`D1534.492,3087.603`,
	`D1523.909,3078.783`,
	`D1515.09,3068.2`,
	`D1506.27,3057.617`,
	`D1499.214,3045.27`,
	`D1492.159,3032.922`,
	`D1485.103,3020.575`,
	`D1481.576,3008.228`,
	`D1400.437,3008.228`,
	`D1400.437,3008.228`,
	`D1393.381,3009.992`,
	`D1386.326,3009.992`,
	`D1047.659,3009.992`,
	`D1047.659,1992.228`,
	`D1045.895,1992.228`,
	`D1015.909,2061.019`,
	`D59.881,2061.019`,
	`D31.659,1992.228`,
	`D29.895,1143.797`,
	`D29.895,1143.797`,

	`L100,1,40,80`,

	`M2908.562,1147.325`,
	`D2908.562,1992.228`,

	`M3019.687,1914.617`,
	`D3019.687,1216.117`,
	`D3619.409,1214.353`,
	`D3617.645,1914.617`,
	`D3017.923,1916.381`,

	`M2982.645,2925.325`,
	`D2982.645,2045.144`,
	`D3667.034,2045.144`,

	`M3667.034,1083.825`,
	`D2987.937,1083.825`,

	`M2982.645,1069.714`,
	`D2982.645,215.992`,

	`M4231.478,1143.797`,
	`D4231.478,1992.228`,

	`M4683.034,1992.228`,
	`D4683.034,1143.797`,

	`M4877.062,1143.797`,
	`D3741.117,1143.797`,
	`D3739.353,1990.464`,
	`D4877.062,1992.228`,

	`L0`,

	`M4854.131,2061.019`,
	`D4713.02,2061.019`,
	`D4684.798,1992.228`,
	`D4683.034,1992.228`,
	`D4653.048,2061.019`,
	`D4263.228,2061.019`,
	`D4235.006,1992.228`,
	`D4229.715,1992.228`,
	`D4201.492,2061.019`,
	`D3765.812,2061.019`,
	`D3737.589,1992.228`,
	`D3621.173,1918.144`,
	`D3674.09,2046.908`,
	`D4032.159,2387.339`,
	`D4032.159,2837.131`,
	`D3850.478,2927.089`,
	`D3765.812,2927.089`,
	`D3765.812,2930.617`,
	`D3656.451,2930.617`,
	`D3656.451,2930.617`,
	`D3651.159,2942.964`,
	`D3645.867,2957.075`,
	`D3638.812,2969.422`,
	`D3631.756,2981.769`,
	`D3622.937,2992.353`,
	`D3614.117,3002.936`,
	`D3603.534,3011.756`,
	`D3592.951,3020.575`,
	`D3580.603,3029.394`,
	`D3568.256,3036.45`,
	`D3555.909,3041.742`,
	`D3541.798,3047.033`,
	`D3529.451,3052.325`,
	`D3513.576,3054.089`,
	`D3499.464,3055.853`,
	`D3485.353,3057.617`,
	`D3485.353,3057.617`,
	`D3469.478,3055.853`,
	`D3455.367,3054.089`,
	`D3441.256,3052.325`,
	`D3427.145,3047.033`,
	`D3413.034,3041.742`,
	`D3400.687,3036.45`,
	`D3388.339,3029.394`,
	`D3377.756,3020.575`,
	`D3367.173,3011.756`,
	`D3356.59,3002.936`,
	`D3347.77,2992.353`,
	`D3338.951,2981.769`,
	`D3330.131,2969.422`,
	`D3324.839,2957.075`,
	`D3317.784,2942.964`,
	`D3314.256,2930.617`,
	`D3076.131,2930.617`,
	`D3076.131,2927.089`,
	`D2980.881,2927.089`,
	`D2980.881,2918.269`,
	`D2848.589,2868.881`,
	`D2848.589,2101.589`,
	`D2980.881,2052.2`,
	`D2980.881,2046.908`,
	`D3021.451,1921.672`,
	`D2910.326,1992.228`,
	`D2848.589,1992.228`,
	`D2848.589,1143.797`,
	`D2910.326,1143.797`,
	`D3012.631,1214.353`,
	`D3016.159,1214.353`,
	`D2980.881,1083.825`,
	`D2980.881,1071.478`,
	`D2848.589,1025.617`,
	`D2848.589,258.325`,
	`D2980.881,212.464`,
	`D2980.881,200.117`,
	`D3076.131,200.117`,
	`D3076.131,198.353`,
	`D3321.312,198.353`,
	`D3321.312,198.353`,
	`D3331.895,175.422`,
	`D3347.77,156.019`,
	`D3365.409,136.617`,
	`D3384.812,120.742`,
	`D3407.742,108.394`,
	`D3432.437,99.575`,
	`D3457.131,92.52`,
	`D3485.353,90.756`,
	`D3485.353,90.756`,
	`D3511.812,92.52`,
	`D3538.27,99.575`,
	`D3561.201,108.394`,
	`D3584.131,120.742`,
	`D3603.534,136.617`,
	`D3621.173,156.019`,
	`D3637.048,175.422`,
	`D3649.395,198.353`,
	`D3765.812,198.353`,
	`D3765.812,200.117`,
	`D3850.478,200.117`,
	`D4032.159,290.075`,
	`D4032.159,738.103`,
	`D3674.09,1080.297`,
	`D3621.173,1207.297`,
	`D3737.589,1142.033`,
	`D3765.812,1073.242`,
	`D4201.492,1073.242`,
	`D4229.715,1142.033`,
	`D4233.242,1142.033`,
	`D4263.228,1073.242`,
	`D4653.048,1073.242`,
	`D4683.034,1142.033`,
	`D4684.798,1142.033`,
	`D4713.02,1073.242`,
	`D4854.131,1073.242`,
	`D4882.353,1142.033`,
	`D4882.353,1142.033`,
	`D4882.353,1142.033`,
	`D4884.117,1143.797`,
	`D4884.117,1990.464`,
	`D4854.131,2061.019`,
	`D4854.131,2061.019`,

	`&1,1,1`,
	`TB50,0`,
	`FO0`,
}

func (r Robo) EasterEgg() {
	for _, cmd := range easteregg {
		r.Printf(cmd)
	}
}
